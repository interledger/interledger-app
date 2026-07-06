package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/interledger/interledger-app/go/backend/providers/pti/external"
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
	}).AnyTimes()

	cl.EXPECT().GetTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, requestID string) (*external.TransactionStatus, error) {
		return &external.TransactionStatus{
			Status: "SETTLED",
		}, nil
	}).AnyTimes()

	cl.EXPECT().CreateTransfer(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, args external.TransferArgs) (*external.IDResponse, error) {
		return &external.IDResponse{
			ID: uuid.NewString(),
		}, nil
	}).AnyTimes()

	encryptedCardInfoPattern := `{"id":"%s","type":"ENCRYPTED_CREDIT_CARD","creditCardLast4":"1234","creditCardBin":"123456","creditCardReference":"reference","creditCardAddress":{"streetAddress":"123 main st","city":"New York","stateCode":"US-NY","country":"US","postalCode":"10005"}}`
	cl.EXPECT().GetUsersPaymentInformation(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, userID, id string) (json.RawMessage, error) {
		raw := fmt.Sprintf(encryptedCardInfoPattern, id)

		return []byte(raw), nil
	}).AnyTimes()

	return cl
}
