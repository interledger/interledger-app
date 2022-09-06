package unit

import (
	"context"
	"encoding/json"
	"net/http"

	"gitlab.com/fynbos/backend/providers/unit/external"
)

type Client interface {
	CreateDepositAccount(ctx context.Context, customerID string) (*DepositAccount, error)
	GetDepositAccount(ctx context.Context, id string) (*DepositAccount, error)
	GetApplicationForm(ctx context.Context, userID string) (*ApplicationForm, error)
	GetStatements(ctx context.Context, customerID string) ([]Statement, error)
	GetStatementPDF(ctx context.Context, args *GetStatementPDFArgs) (*StatementPDF, error)
	CreateApplicationForm(ctx context.Context, args *CreateApplicationFormArgs) (*ApplicationForm, error)
	CreateApplication(ctx context.Context, args *CreateApplicationArgs) (*Application, error)
	CreateCustomer(ctx context.Context, args *CreateCustomerArgs) (*Customer, error)
	GetCustomer(ctx context.Context, id string) (*Customer, error)
	GetCustomerByIdentityID(ctx context.Context, identityID string) (*Customer, error)
	CreateCounterParty(ctx context.Context, args *CreateCounterPartyArgs) (*CounterParty, error)
	GetCounterPartyByFundingsourceID(ctx context.Context, fundingsourceID string) (*CounterParty, error)
	// This will call out to Unit to originate a debit ACH. The result will come via webhook.
	InitiateUserDeposit(ctx context.Context, args *InitiateUserDepositArgs) (*UserAchDeposit, error)
	HandleEvent(ctx context.Context, event external.Event, rawEvent json.RawMessage) error

	VerifyWebhook(ctx context.Context, body []byte, signature string) error
	StoreEvent(ctx context.Context, event external.Event, rawEvent json.RawMessage) (*DbEvent, error)
	GetEvent(ctx context.Context, id string) (*DbEvent, error)
	MakeHttpHandler() http.HandlerFunc
}
