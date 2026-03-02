package v1

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type transactionHandler struct {
	path string
	ctrl *Controller
}

// the requestID is the transaction ID
// https://developers.platform.fiant.io/reference/gettransaction
func (th *transactionHandler) Get(ctx context.Context, requestID string) (*http.Response, error) {
	path, err := url.JoinPath(th.path, requestID)
	if err != nil {
		return nil, err
	}
	return th.ctrl.get(ctx, path)
}

type SandboxActionType string

const (
	SETTLE_ACH SandboxActionType = "SETTLE_ACH"
	RETURN_ACH SandboxActionType = "RETURN_ACH"
)

// only available in sandbox environment, used to simulate settlement and returns for ACH transactions
// https://developers.platform.fiant.io/reference/performaction
func (th *transactionHandler) SandboxAction(ctx context.Context, requestID string, action SandboxActionType) error {
	path, err := url.JoinPath(th.path, requestID, "actions") // "transactions/{requestId}/actions"
	if err != nil {
		return err
	}

	payload := []byte(`{"action":"` + string(action) + `"}`)

	resp, err := th.ctrl.post(ctx, path, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to perform sandbox action, code: %s", resp.Status)
	}

	return nil
}
