package ops

import (
	"context"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/log"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	"go.temporal.io/sdk/client"
	"go.uber.org/zap"
)

var ptiPublicKey = `
-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEAlXgzWngvg4t6oIvQ5/uFiaHT3bdPDyXtN5dK7nBLzA8=
-----END PUBLIC KEY-----
`

func loadEd25519PublicKey() (ed25519.PublicKey, error) {
	keyStr := os.Getenv("PTI_PUBLIC_KEY_JWK")
	if keyStr == "" {
		keyStr = ptiPublicKey
	}
	keyStr = strings.ReplaceAll(keyStr, `\n`, "\n")
	block, _ := pem.Decode([]byte(keyStr))
	if block == nil {
		return nil, fmt.Errorf("pti: failed to decode public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	edKey, ok := pub.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("pti: public key is not Ed25519")
	}
	return edKey, nil
}

func verifyEd25519Signature(key ed25519.PublicKey, body []byte, header string) bool {
	for _, p := range strings.Split(header, ",") {
		p = strings.TrimSpace(p)
		if !strings.HasPrefix(p, "v1=") {
			continue
		}
		sig, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(p, "v1="))
		if err != nil {
			continue
		}
		if ed25519.Verify(key, body, sig) {
			return true
		}
	}
	return false
}

func Webhook(b Backends) (http.HandlerFunc, error) {
	clientID := os.Getenv("PTI_CLIENT_ID")
	pubKey, err := loadEd25519PublicKey()
	if err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("pti webhook: Failed to read body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if !verifyEd25519Signature(pubKey, body, r.Header.Get("X-Signature")) {
			log.Error("pti webhook: signature verification failed")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		var data WebhookData
		if err = json.Unmarshal(body, &data); err != nil {
			log.Error("pti webhook: Failed to unmarshal payload", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if data.ClientID != clientID {
			log.Error("pti webhook: webhook does not match our clientID", zap.String("webhook clientID", data.ClientID))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		log.Info("pti webhook: received webhook", zap.String("payload", string(body)))

		switch data.ResourceType {
		case "USER":
			err = HandleUserUpdate(r.Context(), b, body)
		case "USER_ASSESSMENT", "KYC":
			err = HandleAssessmentUpdate(r.Context(), b, body)
		case "TRANSACTION_STATUS":
			err = HandleTransactionStatus(r.Context(), b, body, w)
		case "TRANSACTION_ASSESSMENT":
			err = HandleTransactionAssessmentUpdate(r.Context(), b, body)
		default:
			log.Error("Unknown pti webhook type", zap.String("externalUserId", data.UserId), zap.String("resourceType", data.ResourceType), zap.String("requestId", data.RequestID))
			w.WriteHeader(http.StatusOK)
			return
		}
		if err != nil {
			log.Error("failed to handle pti webhook", zap.String("externalUserId", data.UserId), zap.String("resourceType", data.ResourceType), zap.String("requestId", data.RequestID), zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}, nil
}

func HandleUserUpdate(ctx context.Context, b Backends, data []byte) error {
	var userData UserWebhook
	if err := json.Unmarshal(data, &userData); err != nil {
		return err
	}
	walletID, err := getWalletID(ctx, b, userData.UserId)
	if err != nil || walletID == "" {
		return fmt.Errorf(" missing user %w %s", pti.ErrInternal, err)
	}

	result, err := b.DB().ExecContext(ctx, "UPDATE pti_users SET status=$1, updated_at=now() WHERE external_id=$2;", userData.Status, userData.UserId)
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	if rows, _ := result.RowsAffected(); rows < 1 {
		return fmt.Errorf("%w Failed to update pti user", pti.ErrInternal)
	}

	return nil
}

func HandleAssessmentUpdate(ctx context.Context, b Backends, data []byte) error {
	var assessmentData AssessmentWebhook
	if err := json.Unmarshal(data, &assessmentData); err != nil {
		return err
	}
	walletID, err := getWalletID(ctx, b, assessmentData.UserId)
	if err != nil || walletID == "" {
		return fmt.Errorf(" missing user %w %s", pti.ErrInternal, err)
	}

	result, err := b.DB().ExecContext(ctx, "UPDATE pti_users SET assessment_status=$1, updated_at=now() WHERE external_id=$2;", assessmentData.Assessment, assessmentData.UserId)
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	if rows, _ := result.RowsAffected(); rows < 1 {
		return fmt.Errorf("%w Failed to update pti user assessment", pti.ErrInternal)
	}

	ptiUser, err := GetUserFromExternalID(ctx, b, assessmentData.UserId)
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	const (
		pendingState           string = "PENDING"
		acceptedState          string = "ACCEPTED"
		refusedState           string = "REFUSED"
		reviewState            string = "UNDER_REVIEW"
		errorState             string = "ERROR"
		requestedMoreInfoState string = "REQUESTED_MORE_INFORMATION"
	)

	switch assessmentData.Assessment {
	case pendingState:
		log.Info("got user in pending state")
	case acceptedState:
		if err := b.KYC().SetKYCStatus(ctx, ptiUser.WalletID, kyc.StatusLevel2); err != nil {
			return fmt.Errorf("%w %s", pti.ErrInternal, err)
		}
	case refusedState:
		if err := b.KYC().SetKYCStatus(ctx, ptiUser.WalletID, kyc.StatusDenied); err != nil {
			return fmt.Errorf("%w %s", pti.ErrInternal, err)
		}
	case reviewState:
		if err := b.KYC().SetKYCStatus(ctx, ptiUser.WalletID, kyc.StatusInReview); err != nil {
			return fmt.Errorf("%w %s", pti.ErrInternal, err)
		}
	case errorState:
		if err := b.KYC().SetKYCStatus(ctx, ptiUser.WalletID, kyc.StatusUnknown); err != nil {
			return fmt.Errorf("%w %s", pti.ErrInternal, err)
		}
	case requestedMoreInfoState:
		if err := b.KYC().SetKYCStatus(ctx, ptiUser.WalletID, kyc.StatusDocumentsRequired); err != nil {
			return fmt.Errorf("%w %s", pti.ErrInternal, err)
		}
	default:
		log.Error("failed to handle pti user assessment webhook", zap.String("externalUserId", assessmentData.UserId), zap.String("assessment_status", assessmentData.Assessment))
		slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "wallet-info-bot", fmt.Sprintf("fiant webhook: kyc assessment status=%s walletID=%s", assessmentData.Assessment, ptiUser.WalletID))
	}

	return nil
}

func HandleTransactionStatus(ctx context.Context, b Backends, raw json.RawMessage, w http.ResponseWriter) error {
	var payload pti.TransactionStatusPayload
	err := json.Unmarshal(raw, &payload)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return err
	}

	const (
		authorizedState  string = "AUTHORIZED"
		refusedState     string = "REFUSED"
		errorState       string = "ERROR"
		pendingState     string = "PENDING"
		processingState  string = "PROCESSING"
		chargedBackState string = "CHARGED_BACK"
		canceledState    string = "CANCELED"
		refundedState    string = "REFUNDED"
		capturedState    string = "CAPTURED"
		settledState     string = "SETTLED"
		clearingFunds    string = "CLEARING_FUNDS"
		returnedState    string = "RETURNED"
	)

	walletID, err := getWalletID(ctx, b, payload.UserID)
	if err != nil || walletID == "" {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	switch payload.Status {
	case authorizedState:
		fmt.Println("got authorized transaction status")
	case refusedState, errorState, canceledState:

		switch payload.TransactionType {
		case "DEPOSIT", "FUNDING":
			err = HandleDepositError(ctx, payload, b)
		case "WITHDRAWAL":
			err = HandleWithdrawError(ctx, payload, b)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
	case pendingState, processingState, clearingFunds:
		// not needed
	case chargedBackState:
		// only for cards
	case refundedState:
		fmt.Println("got refunded transaction status") // REFUNDED also won't apply to the type of transactions that you will be doing (ACH and Wires)
	case capturedState:
		// not needed
	case settledState:
		var err error
		switch payload.TransactionType {
		case "DEPOSIT", "FUNDING":
			err = HandleSettleDeposit(ctx, b, payload)
		case "WITHDRAWAL":
			err = HandleSettleWithdraw(ctx, b, payload)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}
	case returnedState:
		err = HandleReturned(ctx, b, payload)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return err
		}

	default:
		log.Error("failed to handle pti transaction status webhook", zap.String("externalUserId", payload.UserID), zap.String("status", payload.Status), zap.String("requestId", payload.RequestID))
		slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "wallet-info-bot", fmt.Sprintf("fiant webhook: transaction status=%s walletID=%s", payload.Status, payload.UserID))
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

func HandleSettleDeposit(ctx context.Context, b Backends, payload pti.TransactionStatusPayload) error {
	wo := client.StartWorkflowOptions{
		ID:                    "pti_settle_deposit_webhook_" + payload.RequestID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		log.Error("Failed to handle pti deposit webhook", zap.Error(err))
		return err
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	// execute workflow if it's not running
	if workflowStatus != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		_, err = b.Temporal().ExecuteWorkflow(ctx, wo, SettleDepositWorkflow, payload)
		if err != nil {
			log.Error("Failed to handle pti deposit webhook", zap.Error(err))
			return err
		}
	}
	return nil
}

func HandleReturned(ctx context.Context, b Backends, payload pti.TransactionStatusPayload) error {
	wo := client.StartWorkflowOptions{
		ID:                    "pti_returned_webhook_" + payload.RequestID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		log.Error("Failed to handle pti returned webhook", zap.Error(err))
		return err
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	// execute workflow if it's not running
	if workflowStatus != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		_, err = b.Temporal().ExecuteWorkflow(ctx, wo, ReturnedWorkflow, payload)
		if err != nil {
			log.Error("Failed to handle pti returned webhook", zap.Error(err))
			return err
		}
	}
	return nil
}

func HandleSettleWithdraw(ctx context.Context, b Backends, payload pti.TransactionStatusPayload) error {
	wo := client.StartWorkflowOptions{
		ID:                    "pti_settle_withdraw_webhook_" + payload.RequestID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		log.Error("Failed to handle pti deposit webhook", zap.Error(err))
		return err
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	// execute workflow if it's not running
	if workflowStatus != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		_, err = b.Temporal().ExecuteWorkflow(ctx, wo, SettleWithdrawWorkflow, payload)
		if err != nil {
			log.Error("Failed to handle pti deposit webhook", zap.Error(err))
			return err
		}
	}
	return nil
}

func HandleWithdrawError(ctx context.Context, payload pti.TransactionStatusPayload, b Backends) error {
	wo := client.StartWorkflowOptions{
		ID:                    "pti_revert_withdraw__webhook_" + payload.RequestID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:

		log.Error("Failed to handle pti deposit webhook", zap.Error(err))
		return err
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	// execute workflow if it's not running
	if workflowStatus != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		_, err = b.Temporal().ExecuteWorkflow(ctx, wo, RevertWithdrawWorkflow, payload)
		if err != nil {
			log.Error("Failed to handle pti deposit webhook", zap.Error(err))
			return err
		}
	}
	return nil
}

func HandleDepositError(ctx context.Context, payload pti.TransactionStatusPayload, b Backends) error {
	wo := client.StartWorkflowOptions{
		ID:                    "pti_deposit_webhook_mark_transaction_failed" + payload.RequestID,
		TaskQueue:             "backend",
		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
	}

	var workflowStatus enums.WorkflowExecutionStatus
	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
	switch err.(type) {
	case *serviceerror.Internal,
		*serviceerror.Unavailable,
		*serviceerror.InvalidArgument:
		log.Error("Failed to handle pti deposit webhook", zap.Error(err))
		return err
	case *serviceerror.NotFound:
		// do nothing
	default:
		if wflow != nil {
			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
		}
	}

	// execute workflow if it's not running
	if workflowStatus != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
		_, err = b.Temporal().ExecuteWorkflow(ctx, wo, MarkTransactionStateWorkflow, payload, transactions.StateFailed)
		if err != nil {
			log.Error("Failed to handle pti deposit webhook", zap.Error(err))
			return err
		}
	}
	return nil
}

func HandleTransactionAssessmentUpdate(ctx context.Context, b Backends, data []byte) error {
	// TODO
	var payload TransactionAssessmentPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}
	return nil
}

type WebhookData struct {
	ResourceType    string `json:"resourceType"`
	ClientID        string `json:"clientId"`
	RequestID       string `json:"requestId"`
	UserId          string `json:"userId"`
	TransactionType string `json:"transactionType"`
}

type UserWebhook struct {
	ResourceType string `json:"resourceType"`
	ClientID     string `json:"clientId"`
	RequestID    string `json:"requestId"`
	UserId       string `json:"userId"`
	Status       string `json:"status"`
	StatusReason string `json:"statusReason"`
}

type AssessmentWebhook struct {
	ResourceType string `json:"resourceType"`
	ClientID     string `json:"clientId"`
	RequestID    string `json:"requestId"`
	UserId       string `json:"userId"`
	Assessment   string `json:"assessment"`
	Tier         *int   `json:"tier"`
}

type TransactionAssessmentPayload struct {
	ResourceType    string    `json:"resourceType"`
	RequestID       string    `json:"requestId"`
	ClientID        string    `json:"clientId"`
	Amount          int       `json:"amount"`
	Risk            string    `json:"risk"`
	TransactionType string    `json:"transactionType"`
	Assessment      string    `json:"assessment"`
	UserID          string    `json:"userId"`
	Date            time.Time `json:"date"`
}
