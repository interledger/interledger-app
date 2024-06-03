package ops

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go.temporal.io/api/enums/v1"
	temporal_client "go.temporal.io/sdk/client"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	"gitlab.com/fynbos/env"
)

func CheckCredentialsHandler(b Backends) http.HandlerFunc {
	// Start playwright downloads
	go initialSetup()

	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Error("failed to read body", zap.Error(err))
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}

		var ccr CredentialsCheckRequest
		err = json.Unmarshal(body, &ccr)
		if err != nil {
			log.Error("failed to unmarshal body", zap.Error(err))
			http.Error(w, "failed to unmarshal body", http.StatusInternalServerError)
			return
		}

		session, err := Login(ccr.Username, ccr.Password)
		if err != nil {
			log.Error("failed to login", zap.Error(err))
			http.Error(w, "failed to login", http.StatusInternalServerError)
			return
		}

		resp := CredentialsCheckResponse{
			HasMFA:           session.hasMFA,
			CredentialsValid: session.credentialsValid,
		}

		respBytes, err := json.Marshal(resp)
		if err != nil {
			log.Error("failed to create response", zap.Error(err))
			http.Error(w, "failed to create response", http.StatusInternalServerError)
			return
		}

		if resp.CredentialsValid && !resp.HasMFA {
			err = saveWealthUser(req.Context(), b, ccr)
			if err != nil {
				log.Error("failed to save wealth user", zap.Error(err))
				http.Error(w, "failed to save wealth user", http.StatusInternalServerError)
				return
			}

			err = b.Vault().StoreSecret(fmt.Sprintf("%s/credentials/%d/ee", env.GetEnv(), ccr.WealthUserID), ccr.Password)
			if err != nil {
				log.Error("failed to save wealth user password", zap.Error(err))
				http.Error(w, "failed to save wealth user password", http.StatusInternalServerError)
				return
			}
		}

		_, err = b.Temporal().ExecuteWorkflow(context.Background(), temporal_client.StartWorkflowOptions{
			ID:                       "user_ee_tfsa_transactions" + strconv.FormatInt(ccr.WealthUserID, 10),
			TaskQueue:                "backend",
			WorkflowExecutionTimeout: time.Minute * 10, // Workflow has 10 minutes to complete
			WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		}, GetUserTFSATransactionsWorkflow, ccr.WealthUserID)
		if err != nil {
			log.Error("failed to start workflow to check user EE transactions")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(respBytes)
	}
}

func saveWealthUser(ctx context.Context, b Backends, ccr CredentialsCheckRequest) error {
	_, err := b.DB().ExecContext(ctx, "INSERT INTO wealth_users (external_id, easy_equities_username) VALUES ($1, $2) ON CONFLICT (external_id) DO NOTHING", ccr.WealthUserID, ccr.Username)
	return err
}
