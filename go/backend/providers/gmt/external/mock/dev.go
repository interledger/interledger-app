package mock

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/golang/mock/gomock"
	"gitlab.com/fynbos/backend/providers/gmt/external"
)

func SetupDevMock(t *testing.T) *MockClient {
	ctrl := gomock.NewController(t)
	cl := NewMockClient(ctrl)

	cl.EXPECT().OfacVerification(gomock.Any(), gomock.Any()).Return(&external.WsOfac{Error: 0}, nil).AnyTimes()

	cl.EXPECT().ComplianceCheck(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, args external.ComplianceCheck) (*external.WsResponse, error) {
		if args.Transfer.AmountToReceive == 12.2 {
			return &external.WsResponse{Error: 1000, Message: "Invalid User NonRetryable error"}, nil
		}
		return &external.WsResponse{Error: 0, SenderID: rand.Int31(), ReceiverID: rand.Int31()}, nil
	}).AnyTimes()

	cl.EXPECT().InsertTransaction(gomock.Any(), gomock.Any()).Return(&external.WsResponse{
		Error:            0,
		Message:          "Success",
		Password:         fmt.Sprintf("receipt_%d", rand.Int31()),
		Receipt:          "Receipt",
		Receipt_Error_EN: "English_Success",
		Receipt_License:  "Success",
		Receipt_RTR_EN:   "English_Success",
		Status:           "Success",
	}, nil).AnyTimes()

	cl.EXPECT().UpdateTransactionStatus(gomock.Any(), gomock.Any()).Return(&external.WsResponse{Error: 0}, nil).AnyTimes()

	return cl
}
