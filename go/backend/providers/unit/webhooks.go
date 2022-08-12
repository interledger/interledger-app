package unit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"io"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/types"
	"gitlab.com/fynbos/backend/providers/unit/external"
	"go.temporal.io/sdk/client"
)

var (
	ErrDuplicateEvent = errors.New("unit webhook: duplicate event.") // event already stored in database.
)

type Webhook interface {
	HandleEvent(ctx context.Context, event Event, rawEvent json.RawMessage) error
	StoreEvent(ctx context.Context, event Event, rawEvent json.RawMessage) (*DbEvent, error)
	GetEvent(ctx context.Context, id string) (*DbEvent, error)
	MakeHttpHandler() http.HandlerFunc
}

type WebhookArgs struct {
	Up Service       `validate:"required"`
	Db *sqlx.DB      `validate:"required"`
	Tp client.Client `validate:"required"`
}

func NewWebhook(args *WebhookArgs) (Webhook, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, err
	}
	return &webhook{args.Up, args.Db, args.Tp}, nil
}

type webhook struct {
	up Service
	db *sqlx.DB
	tp client.Client
}

func (wh *webhook) HandleEvent(ctx context.Context, event Event, rawEvent json.RawMessage) error {
	_, err := wh.StoreEvent(ctx, event, rawEvent)
	if err != nil {
		if !errors.Is(err, ErrDuplicateEvent) {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}
	}

	switch event.Type {
	case CUSTOMER_CREATED:
		event := &CustomerCreatedEvent{}
		if err := json.Unmarshal(rawEvent, event); err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}
		err := wh.tp.SignalWorkflow(ctx, "unit_onboarding_"+event.Attributes.Tags.FynbosUserId, "", "onboard-unit-customer-created", event)
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}
	case APPLICATION_DENIED:
		event := &ApplicationDeniedEvent{}
		if err := json.Unmarshal(rawEvent, event); err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}
		err := wh.tp.SignalWorkflow(ctx, "unit_onboarding_"+event.Attributes.Tags.FynbosUserId, "", "onboard-unit-application-denied", event)
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}
	case PAYMENT_CREATED, PAYMENT_CLEARING, PAYMENT_SENT,
		PAYMENT_REJECTED, PAYMENT_RETURNED, PAYMENT_CANCELED, PAYMENT_PENDING_REVIEW:
		event := &external.AchPayment{}
		if err := json.Unmarshal(rawEvent, event); err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}
		err := wh.tp.SignalWorkflow(ctx, "deposit_"+event.Attributes.Tags.DepositID, "", "unit-user-ach-deposit", event.Type)
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}
	default:
		// don't fail as Unit may add new events.
	}

	return nil
}

func (wh *webhook) StoreEvent(ctx context.Context, event Event, rawEvent json.RawMessage) (*DbEvent, error) {
	var storedEvent DbEvent

	err := wh.db.GetContext(ctx, &storedEvent, "INSERT INTO unit_events (id, type, raw_event) VALUES ($1, $2, $3) RETURNING *", event.ID, EventType(event.Type), string(rawEvent))
	if err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"primary\"") {
			return nil, fmt.Errorf("%w %s", ErrDuplicateEvent, err)
		}
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &storedEvent, nil
}

func (wh *webhook) GetEvent(ctx context.Context, id string) (*DbEvent, error) {
	var storedEvent DbEvent

	err := wh.db.GetContext(ctx, &storedEvent, `SELECT * FROM unit_events WHERE id = $1 LIMIT 1`, id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &storedEvent, nil
}

func (wh *webhook) MakeHttpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, span := otel.GetTracerProvider().Tracer("").Start(r.Context(), "webhook-span")
		defer span.End()

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to read body")
			http.Error(w, "Failed to parse payload", 500)
			return
		}

		if err := wh.up.VerifyWebhook(context.Background(), payload, r.Header.Get(SignatureHeader)); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to verify webhook")
			http.Error(w, "Signature didn't match.", 401)
			return
		}

		body := ResponseBody{}
		if err := json.Unmarshal(payload, &body); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to unmarshal json")
			http.Error(w, "Failed to parse payload", 500)
			return
		}

		didFail := false
		for _, rawEvent := range body.Data {
			var event Event
			if err := json.Unmarshal(rawEvent, &event); err != nil {
				didFail = true
				continue
			}

			// TODO: this should not fail. Event must be logged.
			err = wh.HandleEvent(context.Background(), event, rawEvent)
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, "failed to handled event")
				http.Error(w, "Failed to handle event", 500)
				return
			}
		}

		// Handling event must not fail. See TODO above.
		// We therefore know it was an unmarshalling error.
		if didFail {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to unmarshal event")
			http.Error(w, "Failed to parse payload", 500)
			return
		}

		w.WriteHeader(200)
	}
}

const (
	CUSTOMER_CREATED               = EventType("customer.created")
	APPLICATION_AWAITING_DOCUMENTS = EventType("application.awaitingdocuments")
	APPLICATION_DENIED             = EventType("application.denied")

	PAYMENT_CREATED        = EventType("payment.created")
	PAYMENT_CLEARING       = EventType("payment.clearing")
	PAYMENT_SENT           = EventType("payment.sent")
	PAYMENT_REJECTED       = EventType("payment.rejected")
	PAYMENT_RETURNED       = EventType("payment.returned")
	PAYMENT_CANCELED       = EventType("payment.canceled")
	PAYMENT_PENDING_REVIEW = EventType("payment.pendingReview")
)

type DbEvent struct {
	ID        string         `db:"id"`
	Type      string         `db:"type"`
	RawEvent  types.JSONText `db:"raw_event"`
	CreatedAt string         `db:"created_at"`
	UpdatedAt string         `db:"updated_at"`
}
