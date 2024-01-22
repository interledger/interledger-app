package mock

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/pti/external"
)

func SetupDevMock(t *testing.T) *MockClient {

	var ctrl *gomock.Controller
	if t == nil {
		ctrl = gomock.NewController(nil)
	} else {
		ctrl = gomock.NewController(t)
	}
	cl := NewMockClient(ctrl)

	cl.EXPECT().WalletWithdrawal(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, args external.WithdrawalArgs) (string, error) {
		id := uuid.NewString()

		return id, nil
	}).AnyTimes()

	cl.EXPECT().WalletDeposit(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, args external.DepositArgs) (string, error) {
		id := uuid.NewString()

		return id, nil
	}).AnyTimes()

	cl.EXPECT().UpdateTransactionStatus(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, args external.UpdateTxStatusArgs) (string, error) {

		return args.TransactionID, nil
	}).AnyTimes()

	cl.EXPECT().GetWallet(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, userID, id string) (*external.Wallet, error) {

		return &external.Wallet{
			WalletID:  "id",
			Currency:  "USD",
			Reference: "ladidaa",
			Balance:   1000000,
		}, nil
	})

	return cl
}
