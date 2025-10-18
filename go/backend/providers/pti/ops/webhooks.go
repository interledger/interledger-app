package ops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwe"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
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

var ptiPublicKeyJwk = `
{
  "e": "AQAB",
  "kid": "861debeb-98ad-4f9a-a144-351e18093ea9",
  "kty": "RSA",
  "n": "1SG6gvMXxVLH_tQn7N7C5eWnYge7fX9uDdoVQdwQgMwbdDY9XLvBsjHIUZVksEt_t6GY42TeNMfbip8okl44-I_6z6cUzS-5xbiBfE_LEJf4NJKb7zftEj30xsyaT5GXLqdM1FuBE5gq2YBxnKvPVxTvNSsN5I6H9TzlCSVbrG-MMPlVzsai-tW0amy1BlPHIoaExk9HYZQXU6bMqozzy9LXXKh1alo4TzIZKeAU7ID5Vscyvfe7z4MNVvAA1v5FEofNAZBasG5gw6Fiolm9vdcB6Y6kxRWLpsifielF-xs_TfAjh2Ff7rT_tAq6C-s2ETa0kj1WFNEevHPZ3-QfKApn1vULfztrat9Q0knstGq5mGJrNPwzF66E1mG-Nf7q5-IRqGvbtiqggZyPvG6VwIwVi-ZQUQ7wZ2sgIUE_-tLxlTuTP2iWn3fjOy9Oban_AnDydpA5mMPG4N1jxPcXsA9x19wWCgeHNe-AVDG87qW-qBSeh08e6y_6kjwOG-AKctzbpVvpWl9pfDOkZGlgK5QO5n72cKvQ53Dmo1CZkGxXWcUDsd8cwQNgOkbJKl73uzd9Wz3gN7_kZtZRgepSlDSeWAfbrUeg5pyKrXdydVaLOIckgmYQhnFksJNxV7RYNAd1iMXmBazygGpo9cK27b-EePoL6KXOOR36oEOjP1s"
}`

// var ptiPublicKeySource = "https://raw.githubusercontent.com/provenancetech/pti-docs/master/utils/pti-prod-public.jwk"

func ParsePTIPublicKey() (jwk.Key, error) {
	keyToParse := os.Getenv("PTI_PUBLIC_KEY_JWK")
	if keyToParse == "" {
		keyToParse = ptiPublicKeyJwk
	}
	return jwk.ParseKey([]byte(keyToParse), jwk.WithPEM(false))
}

func Webhook(b Backends) (http.HandlerFunc, error) {
	clientID := os.Getenv("PTI_CLIENT_ID")
	ptiPublicKey, err := ParsePTIPublicKey()
	if err != nil {
		return nil, err
	}

	var ptiPrivateKey jwk.Key
	if os.Getenv("PTI_JWK") != "" {
		ptiPrivateKey, err = jwk.ParseKey([]byte(os.Getenv("PTI_JWK")))
		if err != nil {
			return nil, err
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error("pti webhook: Failed to read body", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// hack: add header field if it's missing - otherwise jwe library will fail to parse the message
		var result map[string]any
		err = json.Unmarshal(body, &result)
		if err != nil {
			log.Error("pti webhook: Failed to unmarshal payload", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if result["headers"] == nil {
			headers := map[string]string{}
			headers["alg"] = "RSA-OAEP-256"
			headers["enc"] = "A256CBC-HS512"
			result["header"] = headers
		}

		payloadWithHeader, err := json.Marshal(result)
		if err != nil {
			log.Error("pti webhook: Failed to marshal payload with header", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		d, err := jwe.Decrypt(payloadWithHeader, jwe.WithKey(jwa.RSA_OAEP_256, ptiPrivateKey))
		if err != nil {
			log.Error("pti webhook: Failed to decrypt", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		v, err := jws.Verify(d, jws.WithKey(jwa.RS512, ptiPublicKey))
		if err != nil {
			log.Error("pti webhook: Failed to verify", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		var data WebhookData
		if err = json.Unmarshal(v, &data); err != nil {
			log.Error("pti webhook: Failed to unmarshal json", zap.Error(err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if data.ClientID != clientID {
			log.Error("pti webhook: webhook does not match our clientID", zap.String("webhook clientID", data.ClientID))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		log.Info("pti webhook: received webhook", zap.String("payload", string(v)))

		switch data.ResourceType {
		case "USER":
			err = HandleUserUpdate(r.Context(), b, v)
		case "USER_ASSESSMENT", "KYC":
			err = HandleAssessmentUpdate(r.Context(), b, v)
		case "TRANSACTION_STATUS":
			HandleTransactionStatus(r.Context(), b, v, w)
		case "TRANSACTION_ASSESSMENT":

			err = HandleTransactionAssessmentUpdate(r.Context(), b, v)
		default:
			log.Error("Unknown pti webhook type", zap.String("externalUserId", data.UserId), zap.String("resourceType", data.ResourceType), zap.String("requestId", data.RequestID))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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
		slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "Fynbot", fmt.Sprintf("fiant webhook: kyc assessment status=%s walletID=%s", assessmentData.Assessment, ptiUser.WalletID))
	}

	return nil
}

func HandleTransactionStatus(ctx context.Context, b Backends, raw json.RawMessage, w http.ResponseWriter) {
	var payload pti.TransactionStatusPayload
	err := json.Unmarshal(raw, &payload)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
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
	)

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
			return
		}
	case pendingState, processingState:
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
			err = HandleSettleDeposit(ctx, payload, b)
		case "WITHDRAWAL":
			err = HandleSettleWithdraw(ctx, payload, b)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		log.Error("failed to handle pti transaction status webhook", zap.String("externalUserId", payload.UserID), zap.String("status", payload.Status), zap.String("requestId", payload.RequestID))
		slack.SendToChannel(ctx, slack.ChannelNotifyEvents, "Fynbot", fmt.Sprintf("fiant webhook: transaction status=%s walletID=%s", payload.Status, payload.UserID))
	}

	w.WriteHeader(http.StatusOK)
}

func HandleSettleDeposit(ctx context.Context, payload pti.TransactionStatusPayload, b Backends) error {
	fmt.Println("got settled transaction status")

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
		_, err = b.Temporal().ExecuteWorkflow(ctx, wo, SettleDepositWrokflow, payload)
		if err != nil {
			log.Error("Failed to handle pti deposit webhook", zap.Error(err))
			return err
		}
	}
	return nil
}

func HandleSettleWithdraw(ctx context.Context, payload pti.TransactionStatusPayload, b Backends) error {
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
		_, err = b.Temporal().ExecuteWorkflow(ctx, wo, MarkTransactionStateWrokflow, payload, transactions.StateFailed)
		if err != nil {
			log.Error("Failed to handle pti deposit webhook", zap.Error(err))
			return err
		}
	}
	return nil
}

// NOT NEEDED
// func HandlePending(ctx context.Context, payload pti.TransactionStatusPayload, b Backends) error {
// 	fmt.Println("got pending/processing transaction status") // mark as transactions.StatePending
// 	wo := client.StartWorkflowOptions{
// 		ID:                    "pti_deposit_webhook_mark_transaction_pending" + payload.RequestID,
// 		TaskQueue:             "backend",
// 		WorkflowIDReusePolicy: enums.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
// 	}

// 	var workflowStatus enums.WorkflowExecutionStatus
// 	wflow, err := b.Temporal().DescribeWorkflowExecution(ctx, wo.ID, "")
// 	switch err.(type) {
// 	case *serviceerror.Internal,
// 		*serviceerror.Unavailable,
// 		*serviceerror.InvalidArgument:
// 		log.Error("Failed to handle pti deposit webhook", zap.Error(err))
// 		return err
// 	case *serviceerror.NotFound:
// 		// do nothing
// 	default:
// 		if wflow != nil {
// 			workflowStatus = wflow.GetWorkflowExecutionInfo().Status
// 		}
// 	}

// 	// execute workflow if it's not running
// 	if workflowStatus != enums.WORKFLOW_EXECUTION_STATUS_RUNNING {
// 		_, err = b.Temporal().ExecuteWorkflow(ctx, wo, MarkTransactionStateWrokflow, payload, transactions.StatePending)
// 		if err != nil {
// 			log.Error("Failed to handle pti deposit webhook", zap.Error(err))
// 			return err
// 		}
// 	}
// 	return nil
// }

func HandleTransactionAssessmentUpdate(ctx context.Context, b Backends, data []byte) error {
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
