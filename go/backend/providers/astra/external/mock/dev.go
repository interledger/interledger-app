package mock

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/astra"
	"gitlab.com/fynbos/backend/providers/astra/external"
)

var txStatus map[string]astra.TransferStatus

func SetupDevMock(t *testing.T) *MockClient {
	if txStatus == nil {
		txStatus = make(map[string]astra.TransferStatus)
	}
	var ctrl *gomock.Controller
	if t == nil {
		ctrl = gomock.NewController(nil)
	} else {
		ctrl = gomock.NewController(t)
	}
	cl := NewMockClient(ctrl)

	cl.EXPECT().CardToAccount(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, token string, args external.CardToAccountArgs) (*external.CardToAccountResp, error) {
		id := uuid.NewString()

		if args.Amount == 6.66 {
			txStatus[id] = astra.RoutineStatusFailed
		}

		return &external.CardToAccountResp{ID: id}, nil
	}).AnyTimes()

	cl.EXPECT().AccountToCard(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, token string, args external.AccountToCardArgs) (*external.AccountToCardResp, error) {
		id := uuid.NewString()

		if args.Amount == 6.66 {
			txStatus[id] = astra.RoutineStatusFailed
		}

		return &external.AccountToCardResp{ID: id}, nil
	}).AnyTimes()

	cl.EXPECT().GetTransfer(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, token string, txID string) (*external.Transaction, error) {

		status, ok := txStatus[txID]
		if !ok {
			status = astra.TransferStatusProcessed
		}

		return &external.Transaction{
			ID:     txID,
			Status: string(status),
		}, nil
	}).AnyTimes()

	cl.EXPECT().GetRoutine(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, token string, txID string) (*external.Routine, error) {

		status, ok := txStatus[txID]
		if !ok {
			status = astra.RoutineStatusComplete
		}

		return &external.Routine{
			ID:     txID,
			Status: string(status),
		}, nil
	}).AnyTimes()

	cl.EXPECT().CreateAccessToken(gomock.Any(), gomock.Any(), gomock.Any()).Return(&external.AccessToken{
		AccessToken:  uuid.NewString(),
		RefreshToken: uuid.NewString(),
		TokenType:    "bearer",
		ExpiresIn:    7200,
	}, nil)

	return cl
}
