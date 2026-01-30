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

	"gitlab.com/fynbos/backend/rafiki"
	"gitlab.com/fynbos/backend/wallets"

	"gitlab.com/fynbos/backend/linkedaccounts"

	"gitlab.com/fynbos/env"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/gatehub"
	"gitlab.com/fynbos/log"
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

		w.WriteHeader(http.StatusOK)

		go processWebhook(context.Background(), b, hook)
	}
}

func processWebhook(ctx context.Context, b Backends, hook webhook) {
	var err error

	switch hook.Type {
	case "outgoing_payment.created":
		err = outgoingPaymentCreated(ctx, b, hook)
	case "outgoing_payment.completed":
		err = outgoingPaymentCompleted(ctx, b, hook)
	case "incoming_payment.created":
		err = incomingPaymentCreated(ctx, b, hook)
	case "incoming_payment.completed":
		err = incomingPaymentCompleted(ctx, b, hook)
	case "incoming_payment.expired":
		err = incomingPaymentExpired(ctx, b, hook)
	case "incoming_payment.partial_payment_received":
		err = incomingPaymentPartialPaymentReceived(ctx, b, hook)
	default:
		log.Info("rafiki unsupported webhook type", zap.String("type", hook.Type))
		return
	}

	if err != nil {
		log.Error("failed to handle rafiki webhook", zap.String("hook", hook.Type), zap.Error(err))
	}
}

func getReceiverWalletFromIncomingPayment(ctx context.Context, b Backends, incomingPaymentURL string) (*wallets.Wallet, error) {
	const urlPart = "incoming-payments"
	id := incomingPaymentURL
	if strings.Contains(incomingPaymentURL, urlPart) {
		parts := strings.Split(incomingPaymentURL, "/")
		id = parts[len(parts)-1]
	}

	ip, err := b.External().GetIncomingPayment(ctx, id)
	if err != nil {
		log.Error("failed to get incoming payment", zap.Error(err))
		return nil, err
	}

	walletID, err := LookupWalletID(ctx, b, ip.WalletAddressId)
	if err != nil {
		log.Error("failed to lookup receiver wallet ID from incoming payment", zap.Error(err))
		return nil, err
	}

	return b.Wallets().Get(ctx, walletID)
}

func getAccounts(ctx context.Context, b Backends, op outgoingPaymentData) (*linkedaccounts.LinkedAccount, *linkedaccounts.LinkedAccount, error) {
	senderWallet, err := LookupWalletID(ctx, b, op.WalletAddressID)
	if err != nil {
		log.Error("failed to lookup wallet ID from rafiki payment pointer ID", zap.Error(err))
		return nil, nil, err
	}

	receiverWallet, err := getReceiverWalletFromIncomingPayment(ctx, b, op.Receiver)
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

	receiverAccs, err := b.LinkedAccounts().ListBalances(ctx, receiverWallet.ID)
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

func immediatePayment(ctx context.Context, b Backends, op outgoingPaymentData) error {

	amt, err := strconv.ParseUint(op.DebitAmount.Value, 10, 64)
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

func outgoingPaymentCreated(ctx context.Context, b Backends, hook webhook) error {
	var op outgoingPaymentData
	err := json.Unmarshal(hook.Data, &op)
	if err != nil {
		log.Error("failed to unmarshal rafiki outgoing payment", zap.Error(err))
		return err
	}

	amt, err := strconv.ParseUint(op.DebitAmount.Value, 10, 64)
	if err != nil {
		log.Error("failed to convert rafiki outgoing payment amount", zap.Error(err))
		return err
	}

	senderAcc, receiverAcc, err := getAccounts(ctx, b, op)
	if err != nil {
		if errors.Is(err, rafiki.ErrNotFound) {
			log.Warn("canceling outgoing payment - account not found",
				zap.String("paymentId", op.ID),
				zap.Error(err))
			if cancelErr := b.External().CancelOutgoingPayment(ctx, op.ID, err.Error()); cancelErr != nil {
				log.Error("failed to cancel outgoing payment after account lookup failure",
					zap.String("paymentId", op.ID),
					zap.Error(cancelErr))
				return cancelErr
			}
			return nil
		}
		return err
	}

	if senderAcc.WalletID == receiverAcc.WalletID {
		if err := b.External().CancelOutgoingPayment(ctx, op.ID, "sending wallet cannot be the same as receiving wallet"); err != nil {
			log.Error("failed to cancel outgoing payment",
				zap.String("paymentId", op.ID),
				zap.Error(err))
			return err
		}
		return nil
	}

	// More than 1 USD execute immediately
	if amt >= 100 {
		if senderAcc.Provider == gatehub.ProviderName {
			balance, balanceErr := b.Gatehub().GetBalance(ctx, senderAcc.ID)
			if balanceErr != nil {
				log.Error("failed to get sender balance for immediate payment",
					zap.String("paymentId", op.ID),
					zap.String("senderAccount", senderAcc.ID),
					zap.Error(balanceErr))
				return balanceErr
			}
			requiredAmount := currency.FromUInt64(amt, currency.ParseCurrency(op.DebitAmount.AssetCode))
			if balance.Available.Value < requiredAmount.Value {
				log.Warn("canceling outgoing payment - insufficient balance for immediate payment",
					zap.String("paymentId", op.ID),
					zap.String("senderAccount", senderAcc.ID),
					zap.String("available", balance.Available.Format()),
					zap.String("required", requiredAmount.Format()))
				if cancelErr := b.External().CancelOutgoingPayment(ctx, op.ID, "insufficient balance"); cancelErr != nil {
					log.Error("failed to cancel outgoing payment after balance check failure",
						zap.String("paymentId", op.ID),
						zap.Error(cancelErr))
				}
				return nil
			}
		}

		err = immediatePayment(ctx, b, op)
		if err != nil {
			log.Error("immediate payment failed",
				zap.String("paymentId", op.ID),
				zap.Error(err))
			return err
		}
	} else {
		// Reserve the funds for 26 hours, Cron runs every 24.
		// If reserve fails (e.g., insufficient balance), cancel the payment
		err = reserveTransfer(ctx, b, senderAcc, receiverAcc, op.ID, amt, time.Hour*26)
		if err != nil {
			log.Warn("canceling outgoing payment - failed to reserve funds",
				zap.String("paymentId", op.ID),
				zap.Error(err))
			if cancelErr := b.External().CancelOutgoingPayment(ctx, op.ID, err.Error()); cancelErr != nil {
				log.Error("failed to cancel outgoing payment after reserve failure",
					zap.String("paymentId", op.ID),
					zap.Error(cancelErr))
			}
			return nil
		}

		if b.DB() == nil {
			return fmt.Errorf("%w db not configured", rafiki.ErrInternal)
		}

		_, err = b.DB().ExecContext(ctx, "INSERT INTO rafiki_outgoing_payments(id, event_id, from_wallet, to_wallet, amount, amount_asset) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT DO NOTHING;", op.ID, hook.ID, senderAcc.WalletID, receiverAcc.WalletID, op.DebitAmount.Value, op.DebitAmount.AssetCode)
		if err != nil {
			log.Error("failed to add outgoing payment from rafiki outoing payment hook", zap.Error(err))
			return err
		}
	}

	err = b.External().FundOutgoingPayment(ctx, op.ID)
	if err != nil {
		if strings.Contains(err.Error(), "wrong state") {
			log.Info("rafiki outgoing payment already funded",
				zap.String("paymentId", op.ID))
			return nil
		}
		return err
	}

	return nil
}

func outgoingPaymentCompleted(ctx context.Context, b Backends, hook webhook) error {
	if b.DB() == nil {
		return fmt.Errorf("%w db not configured", rafiki.ErrInternal)
	}

	var op outgoingPaymentData
	err := json.Unmarshal(hook.Data, &op)
	if err != nil {
		log.Error("failed to unmarshal rafiki outgoing payment completed", zap.Error(err))
		return err
	}

	_, err = strconv.ParseUint(op.SentAmount.Value, 10, 64)
	if err != nil {
		log.Error("failed to convert rafiki outgoing payment sent amount", zap.Error(err))
		return err
	}

	senderWalletID, err := LookupWalletID(ctx, b, op.WalletAddressID)
	if err != nil {
		log.Error("failed to lookup sender wallet ID from outgoing payment", zap.Error(err))
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	senderAccs, err := b.LinkedAccounts().ListBalances(ctx, senderWalletID)
	if err != nil {
		log.Error("failed to lookup balance accounts for sender", zap.Error(err))
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	var senderAcc *linkedaccounts.LinkedAccount
	for _, la := range senderAccs {
		if op.SentAmount.AssetCode == la.SendCurrency.String() {
			senderAcc = &la
			break
		}
	}
	if senderAcc == nil {
		return fmt.Errorf("%w failed to find sender account for currency=%s", rafiki.ErrNotFound, op.SentAmount.AssetCode)
	}

	var paymentExists bool
	err = b.DB().QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM rafiki_outgoing_payments WHERE id = $1)", op.ID).Scan(&paymentExists)
	if err != nil {
		log.Error("failed to check if outgoing payment was batched",
			zap.String("paymentId", op.ID),
			zap.Error(err))
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	if paymentExists {
		if senderAcc.Provider == gatehub.ProviderName {
			err = b.Gatehub().FinaliseReserve(ctx, op.ID)
		} else {
			return fmt.Errorf("%w unsupported provider: %s", rafiki.ErrInternal, senderAcc.Provider)
		}
		if err != nil {
			log.Error("failed to finalize provider reserve for outgoing payment",
				zap.String("paymentId", op.ID),
				zap.String("provider", senderAcc.Provider),
				zap.Error(err))
			return err
		}
	}

	if err := b.External().WithdrawOutgoingPaymentLiquidity(ctx, op.ID, 0); err != nil {
		// This can happen if payments are happening on the same Rafiki instance, so it is safe to ignore in that case
		log.Error("failed to withdraw outgoing payment liquidity from rafiki after outgoing payment completion",
			zap.String("paymentId", op.ID),
			zap.Error(err))
	} else {
		log.Info("withdrew outgoing payment liquidity from rafiki after outgoing payment completion",
			zap.String("paymentId", op.ID))
	}

	return nil
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

func incomingPaymentCompleted(ctx context.Context, b Backends, hook webhook) error {
	if b.DB() == nil {
		return fmt.Errorf("%w db not configured", rafiki.ErrInternal)
	}

	var ip incomingPaymentData
	if err := json.Unmarshal(hook.Data, &ip); err != nil {
		log.Error("failed to unmarshal rafiki incoming payment completed", zap.Error(err))
		return err
	}

	receivedAmt, err := strconv.ParseUint(ip.ReceivedAmount.Value, 10, 64)
	if err != nil {
		log.Error("failed to convert rafiki incoming payment received amount", zap.Error(err))
		return err
	}

	receiverWalletID, err := LookupWalletID(ctx, b, ip.WalletAddressID)
	if err != nil {
		log.Error("failed to lookup receiver wallet ID from incoming payment", zap.Error(err))
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	receiverAccs, err := b.LinkedAccounts().ListBalances(ctx, receiverWalletID)
	if err != nil {
		log.Error("failed to lookup balance accounts for receiver", zap.Error(err))
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	var receiverAcc *linkedaccounts.LinkedAccount
	for _, la := range receiverAccs {
		if ip.ReceivedAmount.AssetCode == la.ReceiveCurrency.String() {
			receiverAcc = &la
			break
		}
	}
	if receiverAcc == nil {
		return fmt.Errorf("%w failed to find receiver account for currency=%s", rafiki.ErrNotFound, ip.ReceivedAmount.AssetCode)
	}

	receiverAmt := currency.FromUInt64(receivedAmt, currency.ParseCurrency(ip.ReceivedAmount.AssetCode))

	// Credit the receiver: debit intermediary/ops account, credit receiver's account
	if receiverAcc.Provider == gatehub.ProviderName {
		_, err = b.Gatehub().AssignBalance(ctx, receiverAcc.ID, ip.ID, receiverAmt)
	} else {
		// TODO: Support other providers (xago, pti, chimoney)
		return fmt.Errorf("%w unsupported provider: %s", rafiki.ErrInternal, receiverAcc.Provider)
	}
	if err != nil {
		log.Error("failed to assign balance to receiver for incoming payment",
			zap.String("incomingPaymentId", ip.ID),
			zap.String("provider", receiverAcc.Provider),
			zap.Error(err))
		return err
	}

	if err := b.External().WithdrawIncomingPaymentLiquidity(ctx, ip.ID, 0); err != nil {
		log.Error("failed to withdraw incoming payment liquidity from rafiki after incoming payment completion",
			zap.String("incomingPaymentId", ip.ID),
			zap.Error(err))
	} else {
		log.Info("withdrew incoming payment liquidity from rafiki after incoming payment completion",
			zap.String("incomingPaymentId", ip.ID))
	}

	// TODO This might not be needed
	_, err = b.DB().ExecContext(ctx,
		`UPDATE rafiki_incoming_payments 
		 SET completed = true, received_amount = $1, received_amount_asset = $2, updated_at = now()
		 WHERE payment_id = $3`,
		receivedAmt, ip.ReceivedAmount.AssetCode, ip.ID)
	if err != nil {
		log.Error("failed to update incoming payment record",
			zap.String("incomingPaymentId", ip.ID),
			zap.Error(err))
		return err
	}

	return nil
}

func incomingPaymentExpired(ctx context.Context, b Backends, hook webhook) error {
	if b.DB() == nil {
		return fmt.Errorf("%w db not configured", rafiki.ErrInternal)
	}

	var ip incomingPaymentData
	if err := json.Unmarshal(hook.Data, &ip); err != nil {
		log.Error("failed to unmarshal rafiki incoming payment expired", zap.Error(err))
		return err
	}

	receivedAmt, err := strconv.ParseUint(ip.ReceivedAmount.Value, 10, 64)
	if err != nil {
		log.Error("failed to convert rafiki incoming payment expired received amount", zap.Error(err))
		return err
	}

	receiverWalletID, err := LookupWalletID(ctx, b, ip.WalletAddressID)
	if err != nil {
		log.Error("failed to lookup receiver wallet ID from incoming payment (expired)", zap.Error(err))
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	receiverAccs, err := b.LinkedAccounts().ListBalances(ctx, receiverWalletID)
	if err != nil {
		log.Error("failed to lookup balance accounts for receiver (expired)", zap.Error(err))
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	var receiverAcc *linkedaccounts.LinkedAccount
	for _, la := range receiverAccs {
		if ip.ReceivedAmount.AssetCode == la.ReceiveCurrency.String() {
			receiverAcc = &la
			break
		}
	}
	if receiverAcc == nil {
		return fmt.Errorf("%w failed to find receiver account for currency=%s", rafiki.ErrNotFound, ip.ReceivedAmount.AssetCode)
	}

	receiverAmt := currency.FromUInt64(receivedAmt, currency.ParseCurrency(ip.ReceivedAmount.AssetCode))

	// Credit the receiver: debit intermediary/ops account, credit receiver's account
	if receiverAcc.Provider == gatehub.ProviderName {
		_, err = b.Gatehub().AssignBalance(ctx, receiverAcc.ID, ip.ID, receiverAmt)
	} else {
		// TODO: Support other providers (xago, pti, chimoney)
		return fmt.Errorf("%w unsupported provider: %s", rafiki.ErrInternal, receiverAcc.Provider)
	}
	if err != nil {
		log.Error("failed to assign balance to receiver for expired incoming payment",
			zap.String("incomingPaymentId", ip.ID),
			zap.String("provider", receiverAcc.Provider),
			zap.Error(err))
		return err
	}

	if err := b.External().WithdrawIncomingPaymentLiquidity(ctx, ip.ID, 0); err != nil {
		log.Error("failed to withdraw incoming payment liquidity from rafiki after expired incoming payment",
			zap.String("incomingPaymentId", ip.ID),
			zap.Error(err))
	} else {
		log.Info("withdrew incoming payment liquidity from rafiki after expired incoming payment",
			zap.String("incomingPaymentId", ip.ID))
	}

	// TODO This might not be needed
	_, err = b.DB().ExecContext(ctx,
		`UPDATE rafiki_incoming_payments 
		 SET completed = true, received_amount = $1, received_amount_asset = $2, updated_at = now()
		 WHERE payment_id = $3`,
		receivedAmt, ip.ReceivedAmount.AssetCode, ip.ID)
	if err != nil {
		log.Error("failed to update incoming payment record (expired)",
			zap.String("incomingPaymentId", ip.ID),
			zap.Error(err))
		return err
	}

	return nil
}

// Since multiple outgoing payments can pay into a single incoming payment, we treat partial payments
// the same as completed payments regarding accounting and balances - we credit the receiver
// for each partial payment received.
func incomingPaymentPartialPaymentReceived(ctx context.Context, b Backends, hook webhook) error {
	if b.DB() == nil {
		return fmt.Errorf("%w db not configured", rafiki.ErrInternal)
	}

	var ip incomingPaymentData
	if err := json.Unmarshal(hook.Data, &ip); err != nil {
		log.Error("failed to unmarshal rafiki incoming payment partial_payment_received", zap.Error(err))
		return err
	}

	receivedAmt, err := strconv.ParseUint(ip.ReceivedAmount.Value, 10, 64)
	if err != nil {
		log.Error("failed to convert rafiki incoming payment partial received amount", zap.Error(err))
		return err
	}

	receiverWalletID, err := LookupWalletID(ctx, b, ip.WalletAddressID)
	if err != nil {
		log.Error("failed to lookup receiver wallet ID from incoming payment (partial)", zap.Error(err))
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}

	receiverAccs, err := b.LinkedAccounts().ListBalances(ctx, receiverWalletID)
	if err != nil {
		log.Error("failed to lookup balance accounts for receiver (partial)", zap.Error(err))
		return fmt.Errorf("%w %s", rafiki.ErrInternal, err)
	}
	var receiverAcc *linkedaccounts.LinkedAccount
	for _, la := range receiverAccs {
		if ip.ReceivedAmount.AssetCode == la.ReceiveCurrency.String() {
			receiverAcc = &la
			break
		}
	}
	if receiverAcc == nil {
		return fmt.Errorf("%w failed to find receiver account for currency=%s", rafiki.ErrNotFound, ip.ReceivedAmount.AssetCode)
	}

	receiverAmt := currency.FromUInt64(receivedAmt, currency.ParseCurrency(ip.ReceivedAmount.AssetCode))

	// Credit receiver's account
	if receiverAcc.Provider == gatehub.ProviderName {
		_, err = b.Gatehub().AssignBalance(ctx, receiverAcc.ID, ip.ID, receiverAmt)
	} else {
		// TODO: Support other providers (xago, pti, chimoney)
		return fmt.Errorf("%w unsupported provider: %s", rafiki.ErrInternal, receiverAcc.Provider)
	}
	if err != nil {
		log.Error("failed to assign balance to receiver for partial incoming payment",
			zap.String("incomingPaymentId", ip.ID),
			zap.String("provider", receiverAcc.Provider),
			zap.Error(err))
		return err
	}

	if err := b.External().WithdrawIncomingPaymentLiquidity(ctx, ip.ID, 0); err != nil {
		log.Error("failed to withdraw incoming payment liquidity from rafiki after partial payment",
			zap.String("incomingPaymentId", ip.ID),
			zap.Error(err))
	} else {
		log.Info("withdrew incoming payment liquidity from rafiki after partial payment",
			zap.String("incomingPaymentId", ip.ID))
	}

	// TODO This might not be needed
	_, err = b.DB().ExecContext(ctx,
		`UPDATE rafiki_incoming_payments 
		 SET received_amount = received_amount + $1, received_amount_asset = $2, updated_at = now()
		 WHERE payment_id = $3`,
		receivedAmt, ip.ReceivedAmount.AssetCode, ip.ID)
	if err != nil {
		log.Error("failed to update incoming payment record (partial)",
			zap.String("incomingPaymentId", ip.ID),
			zap.Error(err))
		return err
	}

	return nil
}

func reserveTransfer(
	ctx context.Context,
	b Backends,
	senderAcc, receiverAcc *linkedaccounts.LinkedAccount,
	txID string,
	amt uint64,
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
