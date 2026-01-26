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

type SandboxActionType string

const (
	SETTLE_ACH SandboxActionType = "SETTLE_ACH"
	RETURN_ACH SandboxActionType = "RETURN_ACH"
)

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

	// body, err := io.ReadAll(resp.Body)
	// if err != nil {
	// 	return err
	// }
	// defer resp.Body.Close()

	// fmt.Printf("code: %d\n", resp.StatusCode)
	// fmt.Println(string(body))

	if resp.StatusCode != http.StatusCreated {
		// return fmt.Errorf("failed to perform sandbox action, code: %s, response: %s", resp.Status, string(body))
		return fmt.Errorf("failed to perform sandbox action, code: %s", resp.Status)
	}

	return nil
}
