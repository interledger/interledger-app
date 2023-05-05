package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	"gitlab.com/fynbos/backend/providers/tabapay/workflows"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

func CreateCard(ctx context.Context, b Backends, args tabapay.CreateCardArgs) (tabapay.Await, error) {
	wo := client.StartWorkflowOptions{
		ID:                       "tabapay_create_card_" + args.BasisTheoryTokenID,
		TaskQueue:                "backend",
		WorkflowExecutionTimeout: 2 * time.Minute,
		WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}
	await, err := b.Temporal().ExecuteWorkflow(ctx, wo, workflows.CreateTabapayCardWorkflow, args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return await.Get, nil
}

func PullFromCard(ctx context.Context, b Backends, args PullFromCardArgs) (string, error) {
	session3DS, err := Get3DSSession(ctx, b, args.ThreeDSID)
	if err != nil {
		return "", err
	}

	referenceID := args.ReferenceID
	if len(referenceID) > 15 {
		referenceID = referenceID[:15]
	}

	transactionResponse, err := b.External().CreateTransaction(ctx, external.CreateTransactionArgs{
		ReferenceID: referenceID,
		Type:        external.TransactionTypePull,
		Currency:    args.Amount.Currency.ISO4217(),
		Amount:      fmt.Sprintf("%.2f", args.Amount.Float64()),
		Accounts: external.CreateTransactionAccounts{
			SourceAccountID:      args.ProviderID,
			DestinationAccountID: args.SettlementAccountID,
		},
		PullOptions: external.CreateTransactionPullOptions{
			ThreeDS: external.ThreeDS{
				Version:         session3DS.Version,
				ECI:             session3DS.ECI,
				UCAF:            session3DS.UCAF,
				XID:             session3DS.XID,
				DSTransactionID: session3DS.DsTransactionID,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return transactionResponse.TransactionID, nil
}

func PushToCard(ctx context.Context, b Backends, args PullFromCardArgs) (string, error) {
	referenceID := args.ReferenceID
	if len(referenceID) > 15 {
		referenceID = referenceID[:15]
	}

	transactionResponse, err := b.External().CreateTransaction(ctx, external.CreateTransactionArgs{
		ReferenceID: referenceID,
		Type:        external.TransactionTypePush,
		Currency:    args.Amount.Currency.ISO4217(),
		Amount:      fmt.Sprintf("%.2f", args.Amount.Float64()),
		Accounts: external.CreateTransactionAccounts{
			SourceAccountID:      args.SettlementAccountID,
			DestinationAccountID: args.ProviderID,
		},
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return transactionResponse.TransactionID, nil
}

func GetTransaction(ctx context.Context, b Backends, id string) (*tabapay.Transaction, error) {
	trxResp, err := b.External().RetrieveTransaction(ctx, id)
	if errors.Is(err, external.ErrNotFound) {
		return nil, tabapay.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	floatAmount, err := strconv.ParseFloat(trxResp.AmountUSD, 64)
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return &tabapay.Transaction{
		ID:             id,
		ReferenceID:    trxResp.ReferenceID,
		Status:         trxResp.Status,
		OriginalStatus: trxResp.Originally,
		Amount:         currency.FromFloat64(floatAmount, "USD"),
		ReversalStatus: trxResp.ReversalStatus,
	}, nil
}

func Init3DS(ctx context.Context, b Backends, args tabapay.Init3DSArgs) (*tabapay.Init3DSResponse, error) {
	resp, err := b.External().Init3DS(ctx, external.Init3DSArgs{
		Account: external.Account{
			AccountID: args.CardID,
		},
		Order: external.Order{
			OrderID:  args.IdempotencyKey,
			Currency: args.Amount.Currency.ISO4217(),
			Amount:   args.Amount.FormatAmount(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	// TODO: handle resp.SC
	_, err = update3DSSession(ctx, b, tabapay.ThreeDSSession{
		ID:       resp.ID3DS,
		CardID:   args.CardID,
		OrderID:  args.IdempotencyKey,
		Amount:   args.Amount.Value,
		Currency: args.Amount.Currency.String(),
		InitAt:   time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return &tabapay.Init3DSResponse{
		ID:                  resp.ID3DS,
		JWT:                 resp.JWT,
		DeviceCollectionURL: resp.DeviceCollectionURL,
	}, nil
}

func Lookup3DS(ctx context.Context, b Backends, args tabapay.Lookup3DSArgs) (*tabapay.Lookup3DSResponse, error) {
	resp, err := b.External().Lookup3DS(ctx, external.Lookup3DSArgs{
		ID3DS:                   args.ThreeDSID,
		TransactionMode:         string(args.TransactionMode),
		TransactionType:         "C",
		AuthenticationIndicator: string(args.AuthenticationIndicator),
		ProductCode:             string(args.ProductCode),
		Account: external.Account{
			AccountID: args.CardID,
		},
		Order: external.Order{
			OrderID:  args.IdempotencyKey,
			Currency: args.Amount.Currency.ISO4217(),
			Amount:   args.Amount.FormatAmount(),
		},
		Browser: external.Browser{
			BrowserInfo:   args.BrowserInfo,
			DeviceChannel: args.DeviceChannel,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	// TODO: handle resp.SC
	_, err = update3DSSession(ctx, b, tabapay.ThreeDSSession{
		ID:                     args.ThreeDSID,
		Version:                resp.Version3DS,
		Enrolled:               resp.Enrolled,
		ProcessorTransactionID: resp.ProcessorTransactionID,
		DsTransactionID:        resp.DsTransactionID,
		Status:                 resp.Status,
		ChallengeURL:           resp.ChallengeURL,
		Payload:                resp.Payload,
		ECI:                    resp.ECI,
		UCAF:                   resp.UCAF,
		XID:                    resp.XID,
		LookupAt:               time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return &tabapay.Lookup3DSResponse{
		Version:                resp.Version3DS,
		Enrolled:               resp.Enrolled,
		ProcessorTransactionID: resp.ProcessorTransactionID,
		DsTransactionID:        resp.DsTransactionID,
		Status:                 resp.Status,
		ChallengeURL:           resp.ChallengeURL,
		Payload:                resp.Payload,
		ECI:                    resp.ECI,
		UCAF:                   resp.UCAF,
		XID:                    resp.XID,
	}, nil
}

func Authenticate3DS(ctx context.Context, b Backends, args tabapay.Authenticate3DSArgs) (*tabapay.Authenticate3DSResponse, error) {
	resp, err := b.External().Authenticate3DS(ctx, external.Authenticate3DSArgs{
		ID3DS: args.ThreeDSID,
		JWT:   args.JWT,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	_, err = update3DSSession(ctx, b, tabapay.ThreeDSSession{
		ID:                     args.ThreeDSID,
		Status:                 resp.Status,
		Version:                resp.Version3DS,
		Enrolled:               resp.Enrolled,
		ProcessorTransactionID: resp.ProcessorTransactionID,
		DsTransactionID:        resp.DsTransactionID,
		ECI:                    resp.ECI,
		UCAF:                   resp.UCAF,
		XID:                    resp.XID,
		AuthenticatedAt:        time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return &tabapay.Authenticate3DSResponse{
		Status:                 resp.Status,
		Version3DS:             resp.Version3DS,
		Enrolled:               resp.Enrolled,
		ProcessorTransactionID: resp.ProcessorTransactionID,
		DsTransactionID:        resp.DsTransactionID,
		ECI:                    resp.ECI,
		UCAF:                   resp.UCAF,
		XID:                    resp.XID,
	}, nil
}

func Get3DSSession(
	ctx context.Context, b Backends, id string,
) (*tabapay.ThreeDSSession, error) {
	var session dbThreeDSSession
	err := b.DB().GetContext(ctx, &session, fmt.Sprintf("SELECT %s FROM tabapay_3ds_sessions WHERE id=$1 ORDER BY revision DESC LIMIT 1;", dbThreeDSSessionFields), id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w 3DS session not found (id=%s).", tabapay.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return dbToThreeDSSession(session), nil
}

func update3DSSession(
	ctx context.Context, b Backends, session tabapay.ThreeDSSession,
) (*tabapay.ThreeDSSession, error) {
	var old dbThreeDSSession
	err := b.DB().GetContext(
		ctx,
		&old,
		fmt.Sprintf("SELECT %s FROM tabapay_3ds_sessions WHERE id=$1 ORDER BY revision DESC LIMIT 1;", dbThreeDSSessionFields),
		session.ID,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	merged, noop, err := merge3DSSession(old, session)
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}
	if noop {
		return dbToThreeDSSession(merged), nil
	}

	_, err = b.DB().ExecContext(
		ctx,
		fmt.Sprintf("INSERT INTO tabapay_3ds_sessions (%s) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)", dbThreeDSSessionFields),
		merged.ID,
		merged.CardID,
		merged.OrderID,
		merged.Revision,
		merged.Amount,
		merged.Currency,
		merged.Version,
		merged.Enrolled,
		merged.ProcessorTransactionID,
		merged.DsTransactionID,
		merged.Status,
		merged.ECI,
		merged.UCAF,
		merged.XID,
		merged.ChallengeURL,
		merged.Payload,
		merged.InitAt,
		merged.LookupAt,
		merged.AuthenticatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return dbToThreeDSSession(merged), nil
}

func dbToThreeDSSession(dbSession dbThreeDSSession) *tabapay.ThreeDSSession {
	return &tabapay.ThreeDSSession{
		ID:                     dbSession.ID,
		CardID:                 dbSession.CardID,
		OrderID:                dbSession.OrderID,
		Revision:               dbSession.Revision,
		Amount:                 dbSession.Amount,
		Currency:               dbSession.Currency,
		Version:                dbSession.Version,
		Enrolled:               dbSession.Enrolled,
		ProcessorTransactionID: dbSession.ProcessorTransactionID,
		DsTransactionID:        dbSession.DsTransactionID,
		Status:                 dbSession.Status,
		ECI:                    dbSession.ECI,
		UCAF:                   dbSession.UCAF,
		XID:                    dbSession.XID,
		ChallengeURL:           dbSession.ChallengeURL,
		Payload:                dbSession.Payload,
		InitAt:                 dbSession.InitAt,
		LookupAt:               dbSession.LookupAt.Time,
		AuthenticatedAt:        dbSession.AuthenticatedAt.Time,
	}
}

func merge3DSSession(
	old dbThreeDSSession, new tabapay.ThreeDSSession,
) (merged dbThreeDSSession, noop bool, err error) {
	noop = true

	merged.ID = new.ID
	merged.CardID = old.CardID
	merged.OrderID = old.OrderID
	merged.Revision = old.Revision
	merged.Amount = old.Amount
	merged.Currency = old.Currency
	merged.Version = old.Version
	merged.Enrolled = old.Enrolled
	merged.ProcessorTransactionID = old.ProcessorTransactionID
	merged.DsTransactionID = old.DsTransactionID
	merged.Status = old.Status
	merged.ECI = old.ECI
	merged.UCAF = old.UCAF
	merged.XID = old.XID
	merged.ChallengeURL = old.ChallengeURL
	merged.Payload = old.Payload
	merged.InitAt = old.InitAt
	merged.LookupAt = old.LookupAt
	merged.AuthenticatedAt = old.AuthenticatedAt

	if old.CardID != new.CardID && new.CardID != "" {
		merged.CardID = new.CardID
		noop = false
	}

	if old.OrderID != new.OrderID && new.OrderID != "" {
		merged.OrderID = new.OrderID
		noop = false
	}

	if old.Amount != new.Amount && new.Amount != 0 {
		merged.Amount = new.Amount
		noop = false
	}

	if old.Currency != new.Currency && new.Currency != "" {
		merged.Currency = new.Currency
		noop = false
	}

	if old.Version != new.Version && new.Version != "" {
		merged.Version = new.Version
		noop = false
	}

	if old.Enrolled != new.Enrolled && new.Enrolled != "" {
		merged.Enrolled = new.Enrolled
		noop = false
	}

	if old.ProcessorTransactionID != new.ProcessorTransactionID && new.ProcessorTransactionID != "" {
		merged.ProcessorTransactionID = new.ProcessorTransactionID
		noop = false
	}

	if old.DsTransactionID != new.DsTransactionID && new.DsTransactionID != "" {
		merged.DsTransactionID = new.DsTransactionID
		noop = false
	}

	if old.Status != new.Status && new.Status != "" {
		merged.Status = new.Status
		noop = false
	}

	if old.ECI != new.ECI && new.ECI != "" {
		merged.ECI = new.ECI
		noop = false
	}

	if old.UCAF != new.UCAF && new.UCAF != "" {
		merged.UCAF = new.UCAF
		noop = false
	}

	if old.XID != new.XID && new.XID != "" {
		merged.XID = new.XID
		noop = false
	}

	if old.ChallengeURL != new.ChallengeURL && new.ChallengeURL != "" {
		merged.ChallengeURL = new.ChallengeURL
		noop = false
	}

	if old.Payload != new.Payload && new.Payload != "" {
		merged.Payload = new.Payload
		noop = false
	}

	if !old.InitAt.Equal(new.InitAt) && !new.InitAt.IsZero() {
		merged.InitAt = new.InitAt
		noop = false
	}

	if old.LookupAt.Valid {
		if !old.LookupAt.Time.Equal(new.LookupAt) && !new.LookupAt.IsZero() {
			merged.LookupAt = sql.NullTime{Time: new.LookupAt, Valid: true}
			noop = false
		}
	} else if !new.LookupAt.IsZero() {
		merged.LookupAt = sql.NullTime{Time: new.LookupAt, Valid: true}
		noop = false
	}

	if old.AuthenticatedAt.Valid {
		if !old.AuthenticatedAt.Time.Equal(new.AuthenticatedAt) && !new.AuthenticatedAt.IsZero() {
			merged.AuthenticatedAt = sql.NullTime{Time: new.AuthenticatedAt, Valid: true}
			noop = false
		}
	} else if !new.AuthenticatedAt.IsZero() {
		merged.AuthenticatedAt = sql.NullTime{Time: new.AuthenticatedAt, Valid: true}
		noop = false
	}

	merged.Revision = old.Revision
	if !noop {
		merged.Revision = old.Revision + 1
	}

	return merged, noop, err
}
