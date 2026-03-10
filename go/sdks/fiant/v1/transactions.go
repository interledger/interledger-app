// note(bradu): this is just a stub
// nothing here is tested or guaranteed to be correct, it's just a starting point for development

package v1

import (
	"context"
	"fmt"
	"net/http"
)

// note: transactions do not use transactionID in the paths
// requestID is the identifier
// as such, we will use requestID as the identifier for transactions
// this applies on all actions, unless otherwise noted in the documentation
type transactionsService struct {
	client *Client
}

// stub!
// https://developers.platform.fiant.io/reference/gettransaction
func (ts *transactionsService) Get(ctx context.Context, requestID string) (*http.Response, error) {
	path := fmt.Sprintf("transactions/%v", requestID)
	return ts.client.get(ctx, path)
}

type SandboxActionTypeEnum string

const (
	SETTLE_ACH SandboxActionTypeEnum = "SETTLE_ACH"
	RETURN_ACH SandboxActionTypeEnum = "RETURN_ACH"
)

// stub!
// only available in sandbox environment, used to simulate settlement and returns for ACH transactions
// https://developers.platform.fiant.io/reference/performaction
func (ts *transactionsService) SandboxAction(ctx context.Context, requestID string, action SandboxActionTypeEnum) error {
	path := fmt.Sprintf("transactions/%v/actions", requestID)

	payload := []byte(`{"action":"` + string(action) + `"}`)
	resp, err := ts.client.post(ctx, path, payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to perform sandbox action, code: %s", resp.Status)
	}

	return nil
}
