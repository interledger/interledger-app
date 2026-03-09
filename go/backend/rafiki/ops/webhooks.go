package ops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

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

		statusCode := processWebhook(r.Context(), b, hook)
		w.WriteHeader(statusCode)
	}
}

func processWebhook(ctx context.Context, b Backends, hook webhook) int {
	switch hook.Type {
	case "incoming_payment.created":
		if err := incomingPaymentCreated(ctx, b, hook); err != nil {
			log.Error("failed to handle incoming_payment.created", zap.Error(err))
		}
		return http.StatusOK

	case "incoming_payment.completed", "incoming_payment.expired":
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
		return startRafikiWorkflow(ctx, b, hook, func() (string, interface{}) {
			var op outgoingPaymentData
			if err := json.Unmarshal(hook.Data, &op); err != nil {
				return "", nil
			}
			wfID := fmt.Sprintf("rafiki_outgoing_payment_created_%s", op.ID)
			return wfID, op
		})

	case "outgoing_payment.completed":
		return startRafikiWorkflow(ctx, b, hook, func() (string, interface{}) {
			var op outgoingPaymentData
			if err := json.Unmarshal(hook.Data, &op); err != nil {
				return "", nil
			}
			wfID := fmt.Sprintf("rafiki_outgoing_payment_completed_%s", op.ID)
			return wfID, op
		})

	case "outgoing_payment.failed":
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
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
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

