package mock

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/currency"
)

func SetupDevMock(t *testing.T) *MockClient {
	var ctrl *gomock.Controller
	if t == nil {
		ctrl = gomock.NewController(nil)
	} else {
		ctrl = gomock.NewController(t)
	}
	cl := NewMockClient(ctrl)

	cl.EXPECT().CreateTransaction(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, amt currency.Amount, idempotencyKey, beneficiaryID string) (string, error) {
		if amt.Value == 666 {
			return "", fmt.Errorf("failure to withdraw")
		}

		return uuid.NewString(), nil
	}).AnyTimes()

	return cl
}
