package dev

import (
	"context"
	"fmt"
	"math/rand"

	"gitlab.com/fynbos/backend/providers/unit/external"
)

var _ external.Client = client{}

type client struct{}

func NewClient(apiToken string) external.Client {
	return &client{}
}

func (c client) GetStatementPDF(ctx context.Context, args *external.GetStatementPDFArgs) ([]byte, error) {
	return []byte{}, nil
}

func (c client) GetStatements(ctx context.Context, customerID string) ([]external.Statement, error) {
	return []external.Statement{}, nil
}

func (c client) FilterApplicationFormsByUserID(ctx context.Context, userID string) ([]external.ApplicationForm, error) {
	return []external.ApplicationForm{
		{
			ID:   fmt.Sprintf("%d", rand.Int31n(10000)),
			Type: "applicationForm",
			Attributes: external.ApplicationFormAttributes{
				Tags: external.ApplicationTags{
					FynbosUserID: userID,
				},
				Url:   "https://application-form.sh/LJ45W6SSGO6VFFNKMLR5WPOSLH6KMSXQZPGXIPG64SLXHD5TCV4GSYXWZVUSNUEIW2KP5SZOI4RMP6IJRKLF5TTDJTU4TCLU3LQX2XFDIQAMG7TKSXHCQY3KGZ3RFEBYEQCB3GGYUGIUWBXT2ZEIOVNBG72GGNNJKMFJ6",
				Stage: "EnterIndividualInformation",
				ApplicationDetails: external.ApplicationFormPrefill{
					ApplicationType: "Individual",
					Nationality:     "US",
					Email:           "person@test.dev",
				},
				AllowedApplicationTypes: []string{"Individual"},
			},
		},
	}, nil
}

func (c client) CreateApplicationForm(
	ctx context.Context,
	args *external.CreateApplicationFormArgs,
) (*external.ApplicationForm, error) {
	return &external.ApplicationForm{
		ID:   fmt.Sprintf("%d", rand.Int31n(10000)),
		Type: "applicationForm",
		Attributes: external.ApplicationFormAttributes{
			Tags: external.ApplicationTags{
				FynbosUserID: args.ID,
			},
			Url:   "https://application-form.sh/LJ45W6SSGO6VFFNKMLR5WPOSLH6KMSXQZPGXIPG64SLXHD5TCV4GSYXWZVUSNUEIW2KP5SZOI4RMP6IJRKLF5TTDJTU4TCLU3LQX2XFDIQAMG7TKSXHCQY3KGZ3RFEBYEQCB3GGYUGIUWBXT2ZEIOVNBG72GGNNJKMFJ6",
			Stage: "EnterIndividualInformation",
			ApplicationDetails: external.ApplicationFormPrefill{
				ApplicationType: "Individual",
				Nationality:     args.Country,
				Email:           args.Email,
			},
			AllowedApplicationTypes: []string{"Individual"},
		},
	}, nil
}

func (c client) CreateApplication(
	ctx context.Context,
	args *external.CreateApplicationArgs,
) (*external.Application, error) {
	return &external.Application{
		ID:   fmt.Sprintf("%d", rand.Int31n(10000)),
		Type: "individualApplication",
		Attributes: external.ApplicationAttributes{
			CreatedAt: "2020-01-14T14:05:04.718Z",
			FullName: &external.FullName{
				First: args.FirstName,
				Last:  args.LastName,
			},
			Ssn: args.Ssn,
			Address: &external.Address{
				Street:     args.Address.Street,
				State:      args.Address.State,
				City:       args.Address.City,
				PostalCode: args.Address.PostalCode,
				Country:    args.Address.Country,
			},
			DateOfBirth: args.DateOfBirth,
			Email:       args.Email,
			Phone: &external.Phone{
				CountryCode: args.Phone.CountryCode,
				Number:      args.Phone.Number,
			},
			Status: "Approved",
			IP:     "127.0.0.1",
			Tags: &external.ApplicationTags{
				FynbosUserID: args.UserID,
			},
		},
	}, nil
}

func (c client) CreateCounterparty(
	ctx context.Context,
	args *external.CreateCounterpartyArgs,
) (*external.Counterparty, error) {
	return &external.Counterparty{
		ID:   fmt.Sprintf("%d", rand.Int31n(10000)),
		Type: "achCounterparty",
		Relationships: external.CounterpartyRelationships{
			Customer: external.Relationship{
				Data: external.TypeData{
					ID:   args.UnitCustomerID,
					Type: "customer",
				},
			},
		},
	}, nil
}

func (c client) OriginateAch(ctx context.Context, args *external.OriginateAchArgs) (*external.AchPayment, error) {
	return &external.AchPayment{
		ID:   fmt.Sprintf("%d", rand.Int31n(10000)),
		Type: "achPayment",
		Attributes: external.AchPaymentAttributes{
			CreatedAt:   "2020-01-13T16:01:19.346Z",
			Status:      external.AchStatusPending,
			Description: "Funding",
			Direction:   "Debit",
			Amount:      10000,
			Tags: external.DepositTags{
				DepositID: "",
			},
		},
		Relationships: external.AchPaymentRelationships{
			Customer: &external.Relationship{
				Data: external.TypeData{
					ID:   fmt.Sprintf("%d", rand.Int31n(10000)),
					Type: "customer",
				},
			},
			Counterparty: external.Relationship{
				Data: external.TypeData{
					ID:   args.CounterpartyID,
					Type: "counterparty",
				},
			},
			Account: external.Relationship{
				Data: external.TypeData{
					ID:   args.DepositAccountID,
					Type: "depositAccount",
				},
			},
		},
	}, nil
}

func (c client) CreateDepositAccount(
	ctx context.Context,
	args *external.CreateDepositAccountArgs,
) (*external.DepositAccount, error) {
	return &external.DepositAccount{
		ID:   fmt.Sprintf("%d", rand.Int31n(10000)),
		Type: "depositAccount",
		Attributes: external.DepositAccountAttributes{
			CreatedAt:        "2000-05-11T10:19:30.409Z",
			Name:             "Peter parker",
			Status:           "Open",
			DepositProduct:   "checking",
			RoutingNumber:    "812345678",
			AccountNumber:    "1000000002",
			Currency:         "USD",
			BalanceInCents:   10000,
			HoldInCents:      1000,
			AvailableInCents: 9000,
		},
		Relationships: &external.DepositAccountRelationships{
			Customer: external.Relationship{
				Data: external.TypeData{
					ID:   args.CustomerID,
					Type: "customer",
				},
			},
		},
	}, nil
}
