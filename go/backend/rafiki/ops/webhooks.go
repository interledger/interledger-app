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

		switch hook.Type {
		case "outgoing_payment.created":
			err = outgoingPayment(r.Context(), b, hook)
		default:
			log.Info("rafiki unsupported webhook type", zap.String("type", hook.Type))
			w.WriteHeader(http.StatusOK)
			return
		}
		if err != nil {
			log.Error("failed to handle rafiki webhook", zap.String("hook", hook.Type), zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
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

func outgoingPayment(ctx context.Context, b Backends, hook webhook) error {
	var op outgoingPaymentData
	err := json.Unmarshal(hook.Data, &op)
	if err != nil {
		log.Error("failed to unmarshal rafiki outgoing payment", zap.Error(err))
		return err
	}

	var amt uint64
	amt, err = strconv.ParseUint(op.DebitAmount.Value, 10, 64)
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
