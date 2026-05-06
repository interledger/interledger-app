package ops

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/backend/rafiki"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/log"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

type webhook struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type outgoingPaymentData struct {
	ID              string    `json:"id"`
	WalletAddressID string    `json:"walletAddressId"`
	State           string    `json:"state"`
	Receiver        string    `json:"receiver"`
	DebitAmount     amount    `json:"debitAmount"`
	ReceiveAmount   amount    `json:"receiveAmount"`
	SentAmount      amount    `json:"sentAmount"`
	StateAttempts   int       `json:"stateAttempts"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	Balance         string    `json:"balance"`
}

type amount struct {
	Value      string `json:"value"`
	AssetCode  string `json:"assetCode"`
	AssetScale int    `json:"assetScale"`
}

type incomingPaymentData struct {
	ID              string    `json:"id"`
	WalletAddressID string    `json:"walletAddressId"`
	CreatedAt       time.Time `json:"createdAt"`
	ExpiresAt       time.Time `json:"expiresAt"`
	IncomingAmount  *amount   `json:"incomingAmount,omitempty"`
	ReceivedAmount  amount    `json:"receivedAmount"`
	Completed       bool      `json:"completed"`
}

func EventWebhook(b Backends) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		raw, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("failed to read rafiki webhook body", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if !env.IsProd() {
			log.Info("rafiki webhook dump", zap.String("json", string(raw)))
		}

		var hook webhook
		err = json.Unmarshal(raw, &hook)
		if err != nil {
			log.Error("failed to unmarshal rafiki webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 5*time.Minute)
		defer cancel()
		statusCode := processWebhook(ctx, b, hook)
		w.WriteHeader(statusCode)
	}
}

func processWebhook(ctx context.Context, b Backends, hook webhook) int {
	switch hook.Type {
	case "incoming_payment.created":
		isGatehub, err := isGatehubIncomingWebhook(ctx, b, hook)
		if err != nil {
			log.Error("failed to resolve provider for incoming_payment.created", zap.Error(err))
			return http.StatusBadRequest
		}
		if !isGatehub {
			log.Info("skipping incoming_payment.created for non-gatehub provider")
			return http.StatusOK
		}

		if err := incomingPaymentCreated(ctx, b, hook); err != nil {
			log.Error("failed to handle incoming_payment.created", zap.Error(err))
		}
		return http.StatusOK

	case "incoming_payment.completed", "incoming_payment.expired":
		isGatehub, err := isGatehubIncomingWebhook(ctx, b, hook)
		if err != nil {
			log.Error("failed to resolve provider for incoming payment finalized webhook",
				zap.String("type", hook.Type),
				zap.Error(err))
			return http.StatusBadRequest
		}
		if !isGatehub {
			log.Info("skipping incoming payment finalized webhook for non-gatehub provider",
				zap.String("type", hook.Type))
			return http.StatusOK
		}

		return startRafikiWorkflow(ctx, b, hook, func() (string, interface{}) {
			var ip incomingPaymentData
			if err := json.Unmarshal(hook.Data, &ip); err != nil {
				return "", nil
			}
			wfID := fmt.Sprintf("rafiki_incoming_payment_finalized_%s", ip.ID)
			return wfID, RafikiIncomingPaymentFinalizedArgs{
				IncomingPayment: ip,
				WebhookType:     hook.Type,
			}
		})

	case "outgoing_payment.created":
		isGatehubToGatehub, err := isGatehubToGatehubOutgoingWebhook(ctx, b, hook)
		if err != nil {
			log.Error("failed to resolve provider for outgoing_payment.created", zap.Error(err))
			return http.StatusBadRequest
		}

		if !isGatehubToGatehub {
			if err := outgoingPayment(ctx, b, hook); err != nil {
				log.Error("failed to handle outgoing_payment.created for non-gatehub provider", zap.Error(err))
				return http.StatusBadRequest
			}
			return http.StatusOK
		}

		return startRafikiWorkflow(ctx, b, hook, func() (string, interface{}) {
			var op outgoingPaymentData
			if err := json.Unmarshal(hook.Data, &op); err != nil {
				return "", nil
			}
			wfID := fmt.Sprintf("rafiki_outgoing_payment_created_%s", op.ID)
			return wfID, op
		})

	case "outgoing_payment.completed":
		isGatehubToGatehub, err := isGatehubToGatehubOutgoingWebhook(ctx, b, hook)
		if err != nil {
			log.Error("failed to resolve provider for outgoing_payment.completed", zap.Error(err))
			return http.StatusBadRequest
		}
		if !isGatehubToGatehub {
			log.Info("skipping outgoing_payment.completed for non-gatehub provider")
			return http.StatusOK
		}

		return startRafikiWorkflow(ctx, b, hook, func() (string, interface{}) {
			var op outgoingPaymentData
			if err := json.Unmarshal(hook.Data, &op); err != nil {
				return "", nil
			}
			wfID := fmt.Sprintf("rafiki_outgoing_payment_completed_%s", op.ID)
			return wfID, op
		})

	case "outgoing_payment.failed":
		isGatehubToGatehub, err := isGatehubToGatehubOutgoingWebhook(ctx, b, hook)
		if err != nil {
			log.Error("failed to resolve provider for outgoing_payment.failed", zap.Error(err))
			return http.StatusBadRequest
		}
		if !isGatehubToGatehub {
			log.Info("skipping outgoing_payment.failed for non-gatehub provider")
			return http.StatusOK
		}

		return startRafikiWorkflow(ctx, b, hook, func() (string, interface{}) {
			var op outgoingPaymentData
			if err := json.Unmarshal(hook.Data, &op); err != nil {
				return "", nil
			}
			wfID := fmt.Sprintf("rafiki_outgoing_payment_failed_%s", op.ID)
			return wfID, op
		})

	default:
		log.Info("rafiki unsupported webhook type", zap.String("type", hook.Type))
		return http.StatusOK
	}
}

func isGatehubToGatehubOutgoingWebhook(ctx context.Context, b Backends, hook webhook) (bool, error) {
	var op outgoingPaymentData
	if err := json.Unmarshal(hook.Data, &op); err != nil {
		return false, err
	}

	senderAcc, receiverAcc, err := getAccounts(ctx, b, op)
	if err != nil {
		return false, err
	}
	return senderAcc.Provider == gatehub.ProviderName && receiverAcc.Provider == gatehub.ProviderName, nil
}

func isGatehubIncomingWebhook(ctx context.Context, b Backends, hook webhook) (bool, error) {
	var ip incomingPaymentData
	if err := json.Unmarshal(hook.Data, &ip); err != nil {
		return false, err
	}

	assetCode := incomingPaymentAssetCode(ip)
	acc, err := getLinkedAccountByWalletAddressAndAsset(ctx, b, ip.WalletAddressID, assetCode, false)
	if err != nil {
		return false, err
	}
	return acc.Provider == gatehub.ProviderName, nil
}

func incomingPaymentAssetCode(ip incomingPaymentData) string {
	if ip.IncomingAmount != nil && ip.IncomingAmount.AssetCode != "" {
		return ip.IncomingAmount.AssetCode
	}
	return ip.ReceivedAmount.AssetCode
}

func getLinkedAccountByWalletAddressAndAsset(
	ctx context.Context,
	b Backends,
	walletAddressID, assetCode string,
	isSender bool,
) (*linkedaccounts.LinkedAccount, error) {
	walletID, err := LookupWalletID(ctx, b, walletAddressID)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup wallet by payment pointer %s: %w", walletAddressID, err)
	}

	accs, err := b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to list balances for wallet %s: %w", walletID, err)
	}

	for i := range accs {
		acc := accs[i]
		if isSender && acc.SendCurrency.String() == assetCode && acc.Type == "balance" {
			return &acc, nil
		}
		if !isSender && acc.ReceiveCurrency.String() == assetCode && acc.Type == "balance" {
			return &acc, nil
		}
	}

	direction := "receive"
	if isSender {
		direction = "send"
	}
	return nil, fmt.Errorf("%w failed to find linked account for %s asset=%s", rafiki.ErrNotFound, direction, assetCode)
}

func startRafikiWorkflow(ctx context.Context, b Backends, hook webhook, prepare func() (string, interface{})) int {
	wfID, args := prepare()
	if wfID == "" || args == nil {
		log.Error("failed to prepare rafiki workflow args", zap.String("type", hook.Type))
		return http.StatusBadRequest
	}

	var workflowFn interface{}
	switch hook.Type {
	case "incoming_payment.completed", "incoming_payment.expired":
		workflowFn = RafikiIncomingPaymentFinalizedWorkflow
	case "outgoing_payment.created":
		workflowFn = RafikiOutgoingPaymentCreatedWorkflow
	case "outgoing_payment.completed":
		workflowFn = RafikiOutgoingPaymentCompletedWorkflow
	case "outgoing_payment.failed":
		workflowFn = RafikiOutgoingPaymentFailedWorkflow
	default:
		log.Error("no workflow registered for webhook type", zap.String("type", hook.Type))
		return http.StatusBadRequest
	}

	wo := client.StartWorkflowOptions{
		ID:                    wfID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_UNSPECIFIED,
	}

	_, err := b.Temporal().ExecuteWorkflow(ctx, wo, workflowFn, args)
	if err != nil {
		log.Error("failed to start rafiki workflow",
			zap.String("type", hook.Type),
			zap.String("workflowID", wfID),
			zap.Error(err))
		return http.StatusBadRequest
	}

	log.Info("started rafiki workflow",
		zap.String("type", hook.Type),
		zap.String("workflowID", wfID))
	return http.StatusOK
}

func incomingPaymentCreated(ctx context.Context, b Backends, hook webhook) error {
	if b.DB() == nil {
		return fmt.Errorf("%w db not configured", rafiki.ErrInternal)
	}

	var ip incomingPaymentData
	if err := json.Unmarshal(hook.Data, &ip); err != nil {
		log.Error("failed to unmarshal rafiki incoming payment created", zap.Error(err))
		return err
	}

	incomingAsset := ""
	if ip.IncomingAmount != nil {
		incomingAsset = ip.IncomingAmount.AssetCode
	}
	if incomingAsset == "" {
		incomingAsset = ip.ReceivedAmount.AssetCode
	}

	// TODO This might not be needed
	_, err := b.DB().ExecContext(ctx,
		`INSERT INTO rafiki_incoming_payments (payment_id, payment_pointer_id, received_amount, received_amount_asset, completed)
		 SELECT $1, $2, 0, $3, false
		 WHERE NOT EXISTS (SELECT 1 FROM rafiki_incoming_payments WHERE payment_id = $1)`,
		ip.ID, ip.WalletAddressID, incomingAsset)
	if err != nil {
		log.Error("failed to insert incoming payment record",
			zap.String("incomingPaymentId", ip.ID),
			zap.Error(err))
		return err
	}

	return nil
}

func getAccounts(ctx context.Context, b Backends, op outgoingPaymentData) (*linkedaccounts.LinkedAccount, *linkedaccounts.LinkedAccount, error) {
	senderWallet, err := LookupWalletID(ctx, b, op.WalletAddressID)
	if err != nil {
		log.Error("failed to lookup wallet ID from rafiki payment pointer ID", zap.Error(err))
		return nil, nil, err
	}

	receiverWalletID, err := getReceiverWalletIDFromIncomingPayment(ctx, b, op.Receiver)
	if err != nil {
		log.Error("failed to lookup wallet ID from rafiki receiver wallet address", zap.Error(err))
		return nil, nil, err
	}

	senderAccs, err := b.LinkedAccounts().ListBalances(ctx, senderWallet)
	if err != nil {
		log.Error("failed to lookup balance accounts for sender", zap.Error(err))
		return nil, nil, fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	var senderAcc linkedaccounts.LinkedAccount
	for _, la := range senderAccs {
		if op.DebitAmount.AssetCode == la.SendCurrency.String() {
			senderAcc = la
			break
		}
	}
	if senderAcc.ID == "" {
		return nil, nil, fmt.Errorf("%w failed to find sender account for currency=%s", rafiki.ErrNotFound, op.DebitAmount.AssetCode)
	}

	receiverAccs, err := b.LinkedAccounts().ListBalances(ctx, receiverWalletID)
	if err != nil {
		log.Error("failed to lookup balance accounts for receiver", zap.Error(err))
		return nil, nil, err
	}
	var receiverAcc linkedaccounts.LinkedAccount
	for _, la := range receiverAccs {
		if op.DebitAmount.AssetCode == la.ReceiveCurrency.String() {
			receiverAcc = la
			break
		}
	}
	if receiverAcc.ID == "" {
		return nil, nil, fmt.Errorf("%w failed to find receiver account for currency=%s", rafiki.ErrNotFound, op.DebitAmount.AssetCode)
	}

	return &senderAcc, &receiverAcc, nil
}

func getReceiverWalletIDFromIncomingPayment(ctx context.Context, b Backends, receiver string) (string, error) {
	receiverPaymentID := receiver
	if strings.Contains(receiver, "incoming-payments") {
		parts := strings.Split(strings.TrimSuffix(receiver, "/"), "/")
		receiverPaymentID = parts[len(parts)-1]
	}

	ip, err := b.External().GetIncomingPayment(ctx, receiverPaymentID)
	if err != nil {
		return "", err
	}

	walletID, err := LookupWalletID(ctx, b, ip.WalletAddressId)
	if err != nil {
		return "", err
	}

	return walletID, nil
}

func immediatePayment(ctx context.Context, b Backends, op outgoingPaymentData) error {

	amt, err := strconv.ParseInt(op.DebitAmount.Value, 10, 64)
	if err != nil {
		log.Error("failed to convert rafiki outgoing payment amount", zap.Error(err))
		return err
	}

	senderAcc, receiverAcc, err := getAccounts(ctx, b, op)
	if err != nil {
		return err
	}

	p, err := b.Payments().Lookup(ctx, op.ID)
	if errors.Is(err, payments.ErrNotFound) {
		p, err = b.Payments().Create(ctx, payments.CreateArgs{
			IdempotencyKey:  op.ID,
			Sender:          payments.Identity{Type: payments.IdentityTypeWalletID, Identifier: senderAcc.WalletID},
			Receiver:        payments.Identity{Type: payments.IdentityTypeWalletID, Identifier: receiverAcc.WalletID},
			SenderAmount:    currency.FromUInt64(amt, currency.ParseCurrency(op.DebitAmount.AssetCode)),
			SenderAccount:   senderAcc.ID,
			ReceiverAccount: receiverAcc.ID,
			ReceiverAmount:  currency.FromUInt64(amt, currency.ParseCurrency(op.ReceiveAmount.AssetCode)),
			IPAddress:       "41.71.7.104", // TODO: get IP address from somewhere
			Type:            payments.TypeRafikiPeer2Peer,
		})
		if err != nil {
			log.Error("failed to create payment from rafiki outoing payment")
			return err
		}
	} else if err != nil {
		log.Error("failed to lookup existing payment from rafiki outoing payment")
		return err
	}

	if p.State == payments.StateCreated {
		_, _, err = b.Payments().Confirm(ctx, p.ID)
		if err != nil {
			log.Error("failed to confirm payment from rafiki outoing payment")
			return err
		}
	}

	return nil
}

func outgoingPayment(ctx context.Context, b Backends, hook webhook) error {
	var op outgoingPaymentData
	err := json.Unmarshal(hook.Data, &op)
	if err != nil {
		log.Error("failed to unmarshal rafiki outgoing payment", zap.Error(err))
		return err
	}

	var amt int64
	amt, err = strconv.ParseInt(op.DebitAmount.Value, 10, 64)
	if err != nil {
		log.Error("failed to convert rafiki outgoing payment amount", zap.Error(err))
		return err
	}
	// Nothing to do
	if amt == 0 {
		return nil
	}
	senderAcc, receiverAcc, err := getAccounts(ctx, b, op)
	if err != nil {
		return err
	}

	if senderAcc.WalletID == receiverAcc.WalletID {
		err := b.External().CancelOutgoingPayment(ctx, op.ID, "sending wallet cannot be the same as receiving wallet")
		if err != nil {
			log.Error("cannot cancel payment outgoing payment", zap.Error(err))
			return err
		}
		return nil
	}
	// More than 1 USD execute immediately
	if amt >= 100 {
		return immediatePayment(ctx, b, op)
	}

	// Reserve the funds for 26 hours, Cron runs every 24.
	err = reserveTransfer(ctx, b, senderAcc, receiverAcc, op.ID, amt, time.Hour*26)
	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO rafiki_outgoing_payments(id, event_id, from_wallet, to_wallet, amount, amount_asset) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING;", op.ID, hook.ID, senderAcc.WalletID, receiverAcc.WalletID, op.DebitAmount.Value, op.DebitAmount.AssetCode)
	if err != nil {
		log.Error("failed to add outgoing payment from rafiki outoing payment hook", zap.Error(err))
		return err
	}

	// Tell rafiki the payment is successful, we'll actually action it later but the fund are reserved
	err = b.External().FundOutgoingPayment(ctx, op.ID)
	if err != nil {
		return err
	}

	return nil
}

func reserveTransfer(
	ctx context.Context,
	b Backends,
	senderAcc, receiverAcc *linkedaccounts.LinkedAccount,
	txID string,
	amt int64,
	timeout time.Duration,
) error {
	if senderAcc == nil || receiverAcc == nil {
		return fmt.Errorf("%w send and receive accounts not specified when reserving balances", rafiki.ErrInternal)
	}
	if senderAcc.SendCurrency != receiverAcc.ReceiveCurrency {
		return fmt.Errorf("%w send: %s, receive: %s", rafiki.ErrCurrencyNotSupported, senderAcc.SendCurrency, receiverAcc.ReceiveCurrency)
	}

	var err error
	if senderAcc.SendCurrency == currency.EUR {
		_, err = b.Gatehub().ReserveBalance(ctx, senderAcc.ID, txID, currency.FromUInt64(amt, currency.EUR), time.Hour*26)
	} else if senderAcc.SendCurrency == currency.ZAR {
		_, err = b.Xago().ReserveBalance(ctx, senderAcc.ID, txID, currency.FromUInt64(amt, currency.ZAR), time.Hour*26)
	} else if senderAcc.SendCurrency == currency.USD {
		err = b.PTI().ReserveTransfer(ctx, senderAcc.ID, receiverAcc.ID, txID, currency.FromUInt64(amt, currency.USD), time.Hour*26)
	} else if senderAcc.SendCurrency == currency.CAD {
		_, err = b.Chimoney().ReserveBalance(ctx, senderAcc.ID, txID, currency.FromUInt64(amt, currency.CAD), time.Hour*26)
	} else {
		return fmt.Errorf("%w %s", rafiki.ErrCurrencyNotSupported, senderAcc.SendCurrency)
	}
	if err != nil {
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	return nil
}
