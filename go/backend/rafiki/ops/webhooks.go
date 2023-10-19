package ops

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type webhook struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type incomingPaymentData struct {
	Payment struct {
		ID               string `json:"id"`
		PaymentPointerID string `json:"paymentPointerId"`
		CreatedAt        string `json:"createdAt"`
		ExpiresAt        string `json:"expiresAt"`
		ReceivedAmount   amount `json:"receivedAmount"`
		Completed        bool   `json:"completed"`
	} `json:"incomingPayment"`
}

type outgoingPaymentData struct {
	Payment struct {
		ID               string `json:"id"`
		PaymentPointerID string `json:"paymentPointerId"`
		State            string `json:"state"`
		Receiver         string `json:"receiver"`
		CreatedAt        string `json:"createdAt"`
		DebitAmount      amount `json:"debitAmount"`
		ReceiveAmount    amount `json:"receiveAmount"`
		Quote            struct {
			IncomingPaymentID string `json:"receiver"`
		} `json:"quote"`
	} `json:"payment"`
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

		var hook webhook
		err = json.Unmarshal(raw, &hook)
		if err != nil {
			log.Error("failed to unmarshal rafiki webhook", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch hook.Type {
		case "incoming_payment.created", "incoming_payment.completed", "incoming_payment.expired":
			err = incomingPaymentHandle(r.Context(), b, hook.Type, hook.Data)
		case "outgoing_payment.completed", "outgoing_payment.failed":
			err = outgoingPaymentCompleteHandle(r.Context(), b, hook)
		case "outgoing_payment.created":
			err = outgoingPaymentCreatedHandle(r.Context(), b, hook)
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

func extractWalletURL(receiver string) string {
	const urlPart = "incoming-payments"
	if strings.Contains(receiver, urlPart) {
		return receiver[:strings.Index(receiver, urlPart)]
	}

	return receiver
}

func outgoingPaymentCreatedHandle(ctx context.Context, b Backends, hook webhook) error {
	var op outgoingPaymentData
	err := json.Unmarshal(hook.Data, &op)
	if err != nil {
		log.Error("failed to unmarshal rafiki outgoing payment", zap.Error(err))
		return err
	}

	var amt uint64
	amt, err = strconv.ParseUint(op.Payment.DebitAmount.Value, 10, 64)
	if err != nil {
		log.Error("failed to convert rafiki outgoing payment amount", zap.Error(err))
		return err
	}
	// Nothing to do
	if amt == 0 {
		return nil
	}

	typ := payments.TypeRafikiPeer2Peer
	receiverID := payments.Identity{Type: payments.IdentityTypeWalletURL, Identifier: extractWalletURL(op.Payment.Receiver)}
	_, err = b.Wallets().GetFromAddress(ctx, receiverID.Identifier)
	if errors.Is(err, wallets.ErrNoWalletFound) {
		typ = payments.TypeRafiki2External
		receiverID.Type = payments.IdentityTypeExternalWalletURL
	} else if err != nil {
		return err
	}

	senderWallet, err := LookupWalletID(ctx, b, op.Payment.PaymentPointerID)
	if err != nil {
		log.Error("failed to lookup wallet ID from rafiki payment pointer ID", zap.Error(err))
		return err
	}

	senderAcc, err := b.LinkedAccounts().GetDefaultSend(ctx, senderWallet)
	if err != nil {
		log.Error("failed to lookup default send account", zap.Error(err))
		return err
	}

	p, err := b.Payments().Create(ctx, payments.CreateArgs{
		IdempotencyKey: op.Payment.ID,
		Sender:         payments.Identity{Type: payments.IdentityTypeWalletID, Identifier: senderWallet},
		Receiver:       receiverID,
		SenderAmount:   currency.FromUInt64(amt, currency.ParseCurrency(op.Payment.DebitAmount.AssetCode)),
		SenderAccount:  senderAcc.ID,
		ReceiverAmount: currency.FromUInt64(amt, currency.ParseCurrency(op.Payment.ReceiveAmount.AssetCode)),
		IPAddress:      "41.71.7.104", // TODO: get IP address from somewhere
		Type:           typ,
	})
	if errors.Is(err, payments.ErrIdempotencyViolation) {
		p, err = b.Payments().Lookup(ctx, op.Payment.ID)
		if err != nil {
			log.Error("failed to lookup existing payment from rafiki outoing payment")
			return err
		}
	} else if err != nil {
		log.Error("failed to create payment from rafiki outoing payment")
		return err
	}

	if p.State == payments.StateCreated {
		_, _, err = b.Payments().Confirm(ctx, p.ID)
		if err != nil {
			log.Error("failed to confirm payment from rafiki outoing payment")
			return err
		}
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO rafiki_outgoing_payments(id, payment_id, event_id) VALUES ($1, $2, $3)", op.Payment.ID, p.ID, hook.ID)
	if err != nil {
		log.Error("failed to add outgoing payment from rafiki outoing payment hook")
	}

	_, err = b.DB().ExecContext(ctx, "UPDATE rafiki_incoming_payments SET payment_id=$1 WHERE id = $2", p.ID, op.Payment.Quote.IncomingPaymentID)
	if err != nil {
		log.Error("failed to confirm payment from rafiki outoing payment hook")
	}

	return nil
}

func outgoingPaymentCompleteHandle(ctx context.Context, b Backends, hook webhook) error {
	var op outgoingPaymentData
	err := json.Unmarshal(hook.Data, &op)
	if err != nil {
		log.Error("failed to unmarshal rafiki outgoing payment", zap.Error(err))
		return err
	}

	var paymentID string
	err = b.DB().GetContext(ctx, &paymentID, "SELECT payment_id FROM rafiki_outgoing_payments WHERE id=$1", op.Payment.ID)
	if err != nil {
		return err
	}

	success := strings.EqualFold(hook.Type, "outgoing_payment.completed")
	err = b.Payments().SignalExternalPayoutComplete(ctx, paymentID, success)
	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx, "UPDATE rafiki_outgoing_payments SET completed=true WHERE id=$1", op.Payment.ID)
	return err
}

func incomingPaymentHandle(ctx context.Context, b Backends, hookType string, data json.RawMessage) error {
	var payment incomingPaymentData
	err := json.Unmarshal(data, &payment)
	if err != nil {
		log.Error("failed to unmarshal rafiki incoming payment", zap.Error(err))
		return err
	}

	completed := payment.Payment.Completed || hookType == "incoming_payment.completed" || hookType == "incoming_payment.expired"

	var amt uint64
	amt, err = strconv.ParseUint(payment.Payment.ReceivedAmount.Value, 10, 64)
	if err != nil {
		log.Error("failed to convert rafiki incoming payment amount", zap.Error(err))
		return err
	}

	_, err = b.DB().ExecContext(ctx, `INSERT INTO rafiki_incoming_payments 
  (id, payment_pointer_id, completed, received_amount, received_amount_asset) 
	VALUES 
  ($1, $2, $3, $4, $5) ON CONFLICT (id) 
    DO UPDATE SET 
                completed = EXCLUDED.completed, 
                received_amount = EXCLUDED.received_amount,
                updated_at = now()`, payment.Payment.ID, payment.Payment.PaymentPointerID, completed, amt, payment.Payment.ReceivedAmount.AssetCode)
	if err != nil {
		log.Error("failed to upsert rafiki incoming payment", zap.Error(err))
		return err
	}

	return nil
}
