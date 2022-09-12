package client

import (
	"context"
	"encoding/json"

	"gitlab.com/fynbos/backend/providers/unit"
	"gitlab.com/fynbos/backend/providers/unit/external"
	external_client "gitlab.com/fynbos/backend/providers/unit/external/client"
	dev_client "gitlab.com/fynbos/backend/providers/unit/external/client/dev"
	"gitlab.com/fynbos/backend/providers/unit/ops"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var _ unit.Client = client{}

type client struct {
	b            opsBackends
	webhookToken string
}

func NewClient(b Backends, apiToken, webhookToken string) unit.Client {
	ob := opsBackends{
		Backends:     b,
		unitExternal: external_client.NewClient(apiToken),
	}

	if env.IsDev() {
		ob.unitExternal = dev_client.NewClient()
	}

	return &client{
		b:            ob,
		webhookToken: webhookToken,
	}
}

func (c client) CreateDepositAccount(
	ctx context.Context, customerID string,
) (acc *unit.DepositAccount, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to create deposit account",
				zap.String("customerID", customerID),
			)
			return
		}

		log.Debug(
			"Created deposit account",
			zap.String("customerID", customerID),
		)
	}()

	return ops.CreateDepositAccount(ctx, c.b, customerID)
}

func (c client) GetDepositAccount(
	ctx context.Context, id string,
) (acc *unit.DepositAccount, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to get deposit account",
				zap.String("id", id),
			)
			return
		}

		log.Debug(
			"Got deposit account",
			zap.String("id", id),
		)
	}()

	return ops.GetDepositAccount(ctx, c.b, id)
}

func (c client) GetApplicationForm(
	ctx context.Context, userID string,
) (form *unit.ApplicationForm, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to get application form",
				zap.String("userID", userID),
			)
			return
		}

		log.Debug(
			"Got application form",
			zap.String("userID", userID),
		)
	}()

	return ops.GetApplicationForm(ctx, c.b, userID)
}

func (c client) GetStatements(
	ctx context.Context, customerID string,
) (statements []unit.Statement, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to get statements",
				zap.String("customerID", customerID),
			)
			return
		}

		log.Debug(
			"Got statements",
			zap.String("customerID", customerID),
		)
	}()

	return ops.GetStatements(ctx, c.b, customerID)
}

func (c client) GetStatementPDF(
	ctx context.Context, args *unit.GetStatementPDFArgs,
) (ret *unit.StatementPDF, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to get PDF for statement",
				zap.String("customerID", args.CustomerID),
				zap.String("statementID", args.StatementID),
			)
			return
		}

		log.Debug(
			"Got PDF for statement",
			zap.String("customerID", args.CustomerID),
			zap.String("statementID", args.StatementID),
		)
	}()

	return ops.GetStatementPDF(ctx, c.b, args)
}

func (c client) CreateApplicationForm(
	ctx context.Context, args *unit.CreateApplicationFormArgs,
) (form *unit.ApplicationForm, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to create application form",
				zap.String("id", args.ID),
				zap.String("country", args.Country),
				// not logging email
			)
			return
		}

		log.Debug(
			"Created application form",
			zap.String("id", args.ID),
		)
	}()

	return ops.CreateApplicationForm(ctx, c.b, args)
}

func (c client) CreateApplication(
	ctx context.Context, args *unit.CreateApplicationArgs,
) (ret *unit.Application, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to create application",
				zap.String("userID", args.UserID),
				zap.String("city", args.City),
				zap.String("state", args.State),
				zap.String("street", args.Street),
				zap.String("street2", args.Street2),
			)
			return
		}

		log.Debug("Created application")
	}()

	return ops.CreateApplication(ctx, c.b, args)
}

func (c client) VerifyWebhook(
	ctx context.Context, body []byte, signature string,
) (err error) {
	defer func() {
		if err != nil {
			log.Error("Failed to verify webhook")
			return
		}

		log.Debug("Verified webhook")
	}()

	return ops.VerifyWebhook(ctx, body, signature, c.webhookToken)
}

func (c client) CreateCustomer(
	ctx context.Context, args *unit.CreateCustomerArgs,
) (ret *unit.Customer, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to create customer",
				zap.String("id", args.ID),
				zap.String("identityID", args.IdentityID),
				zap.String("type", args.Type),
			)
			return
		}

		log.Debug(
			"Created customer",
			zap.String("id", args.ID),
			zap.String("identityID", args.IdentityID),
			zap.String("type", args.Type),
		)
	}()

	return ops.CreateCustomer(ctx, c.b, args)
}

func (c client) GetCustomer(
	ctx context.Context, id string,
) (ret *unit.Customer, err error) {
	defer func() {
		if err != nil {
			log.Error("Failed to get customer")
			return
		}

		log.Debug("Got customer", zap.String("id", id))
	}()

	return ops.GetCustomer(ctx, c.b, id)
}

func (c client) GetCustomerByIdentityID(
	ctx context.Context, identityID string,
) (ret *unit.Customer, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to get customer by identityID",
				zap.String("identityID", identityID),
			)
			return
		}

		log.Debug(
			"Got customer by identityID",
			zap.String("identityID", identityID),
		)
	}()

	return ops.GetCustomerByIdentityID(ctx, c.b, identityID)
}

func (c client) CreateCounterParty(
	ctx context.Context, args *unit.CreateCounterPartyArgs,
) (ret *unit.CounterParty, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to create counterparty",
				zap.String("unitCustomerID", args.UnitCustomerID),
				zap.String("idempotencyKey", args.IdempotencyKey),
				zap.String("type", args.Type),
				zap.String("name", args.Name),
			)
			return
		}

		log.Debug(
			"Created counterparty",
			zap.String("unitCustomerID", args.UnitCustomerID),
			zap.String("idempotencyKey", args.IdempotencyKey),
		)
	}()

	return ops.CreateCounterParty(ctx, c.b, args)
}

func (c client) GetCounterPartyByFundingsourceID(
	ctx context.Context, fundingsourceID string,
) (ret *unit.CounterParty, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to get counterparty by fundingsourceID",
				zap.String("fundingsourceID", fundingsourceID),
			)
			return
		}

		log.Debug(
			"Get counterparty by fundingsourceID",
			zap.String("fundingsourceID", fundingsourceID),
		)
	}()

	return ops.GetCounterPartyByFundingsourceID(ctx, c.b, fundingsourceID)
}

func (c client) InitiateUserDeposit(
	ctx context.Context, args *unit.InitiateUserDepositArgs,
) (ret *unit.UserAchDeposit, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to initiate user ach deposit",
				zap.String("accountID", args.AccountID),
				zap.String("depositID", args.DepositID),
				zap.String("fundingsourceID", args.FundingsourceID),
				zap.Uint64("amount", args.Amount),
			)
			return
		}

		log.Debug(
			"Initiated user ach deposit",
			zap.String("fundingsourceID", args.FundingsourceID),
		)
	}()

	return ops.InitiateUserDeposit(ctx, c.b, args)
}

func (c client) StoreEvent(
	ctx context.Context, event external.Event, rawEvent json.RawMessage,
) (ret *unit.DbEvent, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to store event",
				zap.String("eventID", event.ID),
				zap.String("eventType", string(event.Type)),
			)
			return
		}

		log.Debug(
			"Store event",
			zap.String("eventID", event.ID),
			zap.String("eventType", string(event.Type)),
		)
	}()

	return ops.StoreEvent(ctx, c.b, event, rawEvent)
}

func (c client) GetEvent(
	ctx context.Context, id string,
) (ret *unit.DbEvent, err error) {
	defer func() {
		if err != nil {
			log.Error(
				"Failed to get event",
				zap.String("id", id),
			)
			return
		}

		log.Debug(
			"Got event",
			zap.String("id", id),
		)
	}()

	return ops.GetEvent(ctx, c.b, id)
}
