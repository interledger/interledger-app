package unit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"gitlab.com/fynbos/backend/onboarding"
	"gitlab.com/fynbos/backend/providers/unit"
)

var (
	ErrInternal = errors.New("unit webhook: internal error.")
)

type Webhook interface {
	HandleEvent(ctx context.Context, eventType EventType, rawEvent json.RawMessage) error
	MakeHttpHandler() http.HandlerFunc
}

type WebhookArgs struct {
	Up unit.Service       `validate:"required"`
	Os onboarding.Service `validate:"required"`
}

func NewWebhook(args *WebhookArgs) (Webhook, error) {
	v := validator.New()
	if err := v.Struct(args); err != nil {
		return nil, err
	}
	return &webhook{args.Up, args.Os}, nil
}

type webhook struct {
	up unit.Service
	os onboarding.Service
}

func (wh *webhook) HandleEvent(ctx context.Context, eventType EventType, rawEvent json.RawMessage) error {
	// TODO: log/store event
	switch eventType {
	case CUSTOMER_CREATED:
		event := &CustomerCreatedEvent{}
		if err := json.Unmarshal(rawEvent, event); err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}

		err := wh.os.InitiateUnitCustomerOnboarding(ctx, &onboarding.InitiateUnitCustomerOnboardingArgs{
			IdentityID: event.Attributes.Tags[unit.ApplicationFormUserIDTag],
			Country:    "US",
		})
		if err != nil {
			return fmt.Errorf("%w %s", ErrInternal, err)
		}
	default:
		// don't fail as Unit may add new events.
	}

	return nil
}

func (wh *webhook) MakeHttpHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := wh.up.VerifyWebhook(context.Background(), r); err != nil {
			http.Error(w, "Signature didn't match.", 401)
			return
		}

		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to parse payload", 500)
			return
		}

		body := Body{}
		if err := json.Unmarshal(payload, &body); err != nil {
			http.Error(w, "Failed to parse payload", 500)
			return
		}

		for _, rawEvent := range body.Data {
			var event Event
			if err := json.Unmarshal(rawEvent, &event); err != nil {
				http.Error(w, "Failed to parse payload", 500)
				return
			}

			// TODO: this should not fail. Event must be logged.
			err := wh.HandleEvent(context.Background(), EventType(event.Type), rawEvent)
			if err != nil {
				http.Error(w, "Failed to handle event", 500)
				return
			}
		}

		w.WriteHeader(200)
	}
}

type (
	Body struct {
		Data []json.RawMessage `json:"data"`
	}
	EventType string
	Event     struct {
		ID   string    `json:"id"`
		Type EventType `json:"type"`
	}
	CustomerCreatedEvent struct {
		ID            string                       `json:"id"`
		Type          string                       `json:"type"`
		Attributes    CustomerCreatedAttributes    `json:"attributes"`
		Relationships CustomerCreatedRelationships `json:"relationships"`
	}

	CustomerCreatedAttributes struct {
		CreatedAt string            `json:"createdAt"`
		Tags      map[string]string `json:"tags"`
	}

	CustomerCreatedRelationships struct {
		Customer    Customer    `json:"customer"`
		Application Application `json:"application"`
	}

	Customer struct {
		Data Data `json:"data"`
	}

	Application struct {
		Data Data `json:"data"`
	}

	Data struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}
)

const (
	CUSTOMER_CREATED = EventType("customer.created")
)
