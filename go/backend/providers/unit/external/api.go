package external

import (
	"context"
)

type Client interface {
	CreateDepositAccount(ctx context.Context, args *CreateDepositAccountArgs) (*DepositAccount, error)
	FilterApplicationFormsByUserID(ctx context.Context, userID string) ([]ApplicationForm, error)
	GetStatements(ctx context.Context, customerID string) ([]Statement, error)
	GetStatementPDF(ctx context.Context, args *GetStatementPDFArgs) ([]byte, error)
	CreateApplicationForm(ctx context.Context, args *CreateApplicationFormArgs) (*ApplicationForm, error)
	CreateApplication(ctx context.Context, args *CreateApplicationArgs) (*Application, error)
	CreateCounterparty(ctx context.Context, args *CreateCounterpartyArgs) (*Counterparty, error)
	OriginateAch(ctx context.Context, args *OriginateAchArgs) (*AchPayment, error)
}
