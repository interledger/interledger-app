package ops

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/nyaruka/phonenumbers"
	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/providers/unit/external"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func GetStatementPDF(
	ctx context.Context, b Backends, externalClient external.Client, args *unit.GetStatementPDFArgs,
) (*unit.StatementPDF, error) {
	statement, err := externalClient.GetStatementPDF(ctx, &external.GetStatementPDFArgs{
		ID:         args.StatementID,
		CustomerID: args.CustomerID,
	})

	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			return nil, fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
		}
		if errors.Is(err, external.ErrRequest) {
			return nil, fmt.Errorf("%w %s", unit.ErrClient, err)
		}
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return &unit.StatementPDF{
		ID:  args.StatementID,
		PDF: statement,
	}, nil
}

func GetStatements(
	ctx context.Context, b Backends, externalClient external.Client, customerID string,
) ([]unit.Statement, error) {
	statements, err := externalClient.GetStatements(ctx, customerID)
	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			return nil, fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
		}
		if errors.Is(err, external.ErrRequest) {
			return nil, fmt.Errorf("%w %s", unit.ErrClient, err)
		}
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	if len(statements) == 0 {
		return nil, fmt.Errorf("%w no statements found", unit.ErrNotFound)
	}

	var statementsOut []unit.Statement
	for _, s := range statements {
		statementsOut = append(statementsOut, unit.Statement{
			ID:        s.ID,
			Period:    s.Attributes.Period,
			AccountID: s.Relationships.Account.Data.ID,
		})
	}

	return statementsOut, nil
}

func GetApplicationForm(
	ctx context.Context, b Backends, externalClient external.Client, userID string,
) (*unit.ApplicationForm, error) {
	forms, err := externalClient.FilterApplicationFormsByUserID(ctx, userID)
	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			return nil, fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
		}
		if errors.Is(err, external.ErrRequest) {
			return nil, fmt.Errorf("%w %s", unit.ErrClient, err)
		}
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	if len(forms) != 1 {
		return nil, unit.ErrNotFound
	}

	return &unit.ApplicationForm{
		ID:  forms[0].ID,
		URL: forms[0].Attributes.Url,
	}, nil
}

func CreateApplicationForm(
	ctx context.Context, b Backends, externalClient external.Client, args *unit.CreateApplicationFormArgs,
) (*unit.ApplicationForm, error) {
	form, err := externalClient.CreateApplicationForm(ctx, &external.CreateApplicationFormArgs{
		ID:      args.ID,
		Email:   args.Email,
		Country: args.Country,
	})
	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			return nil, fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
		}
		if errors.Is(err, external.ErrRequest) {
			return nil, fmt.Errorf("%w %s", unit.ErrClient, err)
		}
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return &unit.ApplicationForm{
		ID:  form.ID,
		URL: form.Attributes.Url,
	}, nil
}

func CreateApplication(
	ctx context.Context, b Backends, externalClient external.Client, args *unit.CreateApplicationArgs,
) (application *unit.Application, err error) {
	defer func(begin time.Time) {
		if err != nil {
			log.Error(
				"failed to create unit application",
				zap.String("userId", args.UserID),
				zap.Int64("took", time.Since(begin).Milliseconds()),
				zap.String("msg", err.Error()),
			)
			return
		}

		log.Debug(
			"created unit application",
			zap.String("userId", args.UserID),
			zap.Int64("took", time.Since(begin).Milliseconds()),
		)
	}(time.Now())

	id, err := b.Identity().Get(ctx, args.UserID)
	if err != nil {
		err = fmt.Errorf("%w %s", unit.ErrInternal, err)
		return
	}

	phoneNumber, err := phonenumbers.Parse(id.MobileNumber, "")
	if err != nil {
		err = fmt.Errorf("%w %s", unit.ErrInternal, err)
		return
	}

	// deviceFingerprints := make([]DeviceFingerprint, len(args.DeviceFingerprints))
	// for _, fingerprint := range args.DeviceFingerprints {
	// 	deviceFingerprints = append(deviceFingerprints, DeviceFingerprint{
	// 		Provider: "iovation",
	// 		Value:    fingerprint,
	// 	})
	// }

	app, err := externalClient.CreateApplication(ctx, &external.CreateApplicationArgs{
		UserID:             args.UserID,
		Email:              id.Email,
		IpAddress:          args.IpAddress,
		DateOfBirth:        args.DateOfBirth,
		DeviceFingerprints: nil, // TODO: figure out why this isn't working
		FirstName:          id.FirstName,
		LastName:           id.LastName,
		Ssn:                args.Ssn,
		Address: external.Address{
			Street:     args.Street,
			Street2:    args.Street2,
			City:       args.City,
			State:      args.State,
			PostalCode: args.PostalCode,
			Country:    id.Country,
		},
		Phone: external.Phone{
			CountryCode: strconv.Itoa(int(*phoneNumber.CountryCode)),
			Number:      strconv.Itoa(int(*phoneNumber.NationalNumber)),
		},
	})
	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			err = fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
			return
		}
		if errors.Is(err, external.ErrRequest) {
			err = fmt.Errorf("%w %s", unit.ErrClient, err)
			return
		}
		err = fmt.Errorf("%w %s", unit.ErrInternal, err)
		return
	}

	return &unit.Application{
		Type:         app.Type,
		ID:           app.ID,
		Status:       app.Attributes.Status,
		FynbosUserId: app.Attributes.Tags.FynbosUserID,
		Archived:     app.Attributes.Archived,
		CustomerID:   app.Relationships.Customer.Data.ID,
	}, nil
}

func VerifyWebhook(ctx context.Context, body []byte, signature, webhookToken string) error {
	mac := hmac.New(sha1.New, []byte(webhookToken))
	mac.Write(body)
	sha := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if sha != signature {
		return unit.ErrUnauthorized
	}

	return nil
}

func CreateCustomer(
	ctx context.Context, b Backends, externalClient external.Client, args *unit.CreateCustomerArgs,
) (*unit.Customer, error) {
	id, err := b.Identity().Get(ctx, args.IdentityID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	var customer unit.Customer
	err = b.DB().GetContext(
		ctx,
		&customer,
		"INSERT INTO unit_customers (id, identity_id, type) VALUES ($1, $2, $3) RETURNING *;",
		args.ID,
		id.ID,
		args.Type,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return &customer, nil
}

func GetCustomer(
	ctx context.Context, b Backends, id string,
) (*unit.Customer, error) {
	var ret unit.Customer
	err := b.DB().GetContext(ctx, &ret, "SELECT * FROM unit_customers WHERE id=$1", id)
	if err == sql.ErrNoRows {
		return nil, unit.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return &ret, nil
}

func GetCustomerByIdentityID(
	ctx context.Context, b Backends, identityID string,
) (*unit.Customer, error) {
	var ret unit.Customer
	err := b.DB().GetContext(ctx, &ret, "SELECT * FROM unit_customers WHERE identity_id=$1", identityID)
	if err == sql.ErrNoRows {
		return nil, unit.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return &ret, nil
}

// // This will create the counter party on Unit and store a record of it in our database.
func CreateCounterParty(
	ctx context.Context, b Backends, externalClient external.Client, args *unit.CreateCounterPartyArgs,
) (*unit.CounterParty, error) {
	counterparty, err := externalClient.CreateCounterparty(ctx, &external.CreateCounterpartyArgs{
		Name:           args.Name,
		UnitCustomerID: args.UnitCustomerID,
		RoutingNumber:  args.RoutingNumber,
		AccountNumber:  args.AccountNumber,
		AccountType:    args.AccountType,
		Type:           args.Type,
		IdempotencyKey: args.IdempotencyKey,
	})
	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			return nil, fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
		}
		if errors.Is(err, external.ErrRequest) {
			return nil, fmt.Errorf("%w %s", unit.ErrClient, err)
		}
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	ret := &unit.CounterParty{}
	err = b.DB().GetContext(
		ctx,
		ret,
		"INSERT INTO unit_counterparties (id, fundingsource_id) VALUES ($1, $2) RETURNING *;",
		counterparty.ID,
		args.FundingsourceID,
	)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), unit.ErrInternal)
	}

	return ret, nil
}

func GetCounterPartyByFundingsourceID(
	ctx context.Context, b Backends, fundingsourceID string,
) (*unit.CounterParty, error) {
	ret := &unit.CounterParty{}
	err := b.DB().GetContext(
		ctx,
		ret,
		"SELECT * FROM unit_counterparties WHERE fundingsource_id=$1 LIMIT 1;",
		fundingsourceID,
	)
	if err == sql.ErrNoRows {
		return nil, unit.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return ret, nil
}

func CreateDepositAccount(
	ctx context.Context, b Backends, externalClient external.Client, customerID string,
) (*unit.DepositAccount, error) {
	if customerID == "" {
		return nil, fmt.Errorf("%w CustomerID is required", unit.ErrInvalidArgument)
	}

	idempotencyKey := sha256.Sum256([]byte(customerID))
	depositAccount, err := externalClient.CreateDepositAccount(ctx, &external.CreateDepositAccountArgs{
		CustomerID:     customerID,
		DepositProduct: "checking",
		IdempotencyKey: string(idempotencyKey[0:]),
		Type:           "depositAccount",
	})
	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			return nil, fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
		}
		if errors.Is(err, external.ErrRequest) {
			return nil, fmt.Errorf("%w %s", unit.ErrClient, err)
		}
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	var ret unit.DepositAccount
	err = b.DB().GetContext(
		ctx,
		&ret,
		"INSERT INTO unit_deposit_accounts (id, customer_id) VALUES ($1, $2) RETURNING *;",
		depositAccount.ID,
		customerID,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return &ret, nil
}

func GetDepositAccount(ctx context.Context, b Backends, id string) (*unit.DepositAccount, error) {
	var ret unit.DepositAccount
	err := b.DB().GetContext(
		ctx,
		&ret,
		"SELECT * FROM unit_deposit_accounts WHERE id=$1;",
		id,
	)
	if err == sql.ErrNoRows {
		return nil, unit.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return &ret, nil
}

func IsRetryableError(err error) bool {
	return errors.Is(err, unit.ErrClient) ||
		errors.Is(err, unit.ErrRateLimit) ||
		errors.Is(err, unit.ErrServer) ||
		errors.Is(err, unit.ErrTimeout)
}

func statusToError(statusCode int) error {
	if statusCode == http.StatusRequestTimeout {
		return unit.ErrTimeout
	}
	if statusCode == http.StatusTooManyRequests {
		return unit.ErrRateLimit
	}
	if statusCode == http.StatusNotFound {
		return unit.ErrNotFound
	}
	if statusCode > http.StatusInternalServerError {
		return unit.ErrServer
	} else {
		return unit.ErrInternal
	}
}

func InitiateUserDeposit(
	ctx context.Context, b Backends, externalClient external.Client, args *unit.InitiateUserDepositArgs,
) (*unit.UserAchDeposit, error) {
	acc, err := b.Accounts().Get(ctx, args.AccountID)
	if err != nil {
		return nil, err
	}

	counterparty, err := GetCounterPartyByFundingsourceID(ctx, b, args.FundingsourceID)
	if err != nil {
		return nil, err
	}

	idempotencyKey := sha256.Sum256([]byte(args.DepositID))
	achPayment, err := externalClient.OriginateAch(ctx, &external.OriginateAchArgs{
		IdempotencyKey:   string(idempotencyKey[0:]),
		Amount:           args.Amount,
		Direction:        "Debit",
		CounterpartyID:   counterparty.ID,
		DepositAccountID: acc.ProviderID,
		Description:      args.Description,
	})
	if err != nil {
		var errHttp *external.ErrHttp
		if errors.As(err, &errHttp) {
			return nil, fmt.Errorf("%w %s", statusToError(errHttp.Code), err)
		}
		if errors.Is(err, external.ErrRequest) {
			return nil, fmt.Errorf("%w %s", unit.ErrClient, err)
		}
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	ret := &unit.UserAchDeposit{}
	err = b.DB().GetContext(
		ctx,
		ret,
		`
		INSERT INTO unit_user_ach_deposits (id, deposit_id, deposit_account_id, counterparty_id,
		amount)VALUES ($1, $2, $3, $4, $5) RETURNING *;
		`,
		achPayment.ID,
		args.DepositID,
		acc.ProviderID,
		counterparty.ID,
		args.Amount,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return ret, nil
}

func StoreEvent(ctx context.Context, b Backends, event external.Event, rawEvent json.RawMessage) (*unit.DbEvent, error) {
	var storedEvent unit.DbEvent

	err := b.DB().GetContext(ctx, &storedEvent, "INSERT INTO unit_events (id, type, raw_event) VALUES ($1, $2, $3) RETURNING *", event.ID, external.EventType(event.Type), string(rawEvent))
	if err != nil {
		if strings.Contains(err.Error(), "pq: duplicate key value violates unique constraint \"primary\"") {
			return nil, fmt.Errorf("%w %s", unit.ErrDuplicateEvent, err)
		}
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return &storedEvent, nil
}

func GetEvent(ctx context.Context, b Backends, id string) (*unit.DbEvent, error) {
	var storedEvent unit.DbEvent

	err := b.DB().GetContext(ctx, &storedEvent, `SELECT * FROM unit_events WHERE id = $1 LIMIT 1`, id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", unit.ErrInternal, err)
	}

	return &storedEvent, nil
}

func HandleEvent(ctx context.Context, b Backends, event external.Event, rawEvent json.RawMessage) error {
	_, err := StoreEvent(ctx, b, event, rawEvent)
	if err != nil {
		if err != unit.ErrDuplicateEvent {
			return fmt.Errorf("%w %s", unit.ErrInternal, err)
		}
	}

	switch event.Type {
	case external.CUSTOMER_CREATED:
		event := &external.CustomerCreatedEvent{}
		if err := json.Unmarshal(rawEvent, event); err != nil {
			return fmt.Errorf("%w %s", unit.ErrInternal, err)
		}
		err := b.Temporal().SignalWorkflow(ctx, "unit_onboarding_"+event.Attributes.Tags.FynbosUserID, "", "onboard-unit-customer-created", event)
		if err != nil {
			return fmt.Errorf("%w %s", unit.ErrInternal, err)
		}
	case external.APPLICATION_DENIED:
		event := &external.ApplicationDeniedEvent{}
		if err := json.Unmarshal(rawEvent, event); err != nil {
			return fmt.Errorf("%w %s", unit.ErrInternal, err)
		}
		err := b.Temporal().SignalWorkflow(ctx, "unit_onboarding_"+event.Attributes.Tags.FynbosUserID, "", "onboard-unit-application-denied", event)
		if err != nil {
			return fmt.Errorf("%w %s", unit.ErrInternal, err)
		}
	case external.PAYMENT_CREATED, external.PAYMENT_CLEARING, external.PAYMENT_SENT,
		external.PAYMENT_REJECTED, external.PAYMENT_RETURNED, external.PAYMENT_CANCELED, external.PAYMENT_PENDING_REVIEW:
		event := &external.AchPayment{}
		if err := json.Unmarshal(rawEvent, event); err != nil {
			return fmt.Errorf("%w %s", unit.ErrInternal, err)
		}
		err := b.Temporal().SignalWorkflow(ctx, "deposit_"+event.Attributes.Tags.DepositID, "", "unit-user-ach-deposit", event.Type)
		if err != nil {
			return fmt.Errorf("%w %s", unit.ErrInternal, err)
		}
	default:
		// don't fail as Unit may add new events.
	}

	return nil
}
