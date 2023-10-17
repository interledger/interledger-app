package mock

import (
	context "context"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
)

var txIDAmounts map[string]string

func SetupDevMock(t *testing.T) *MockClient {
	txIDAmounts = make(map[string]string)
	var ctrl *gomock.Controller
	if t == nil {
		ctrl = gomock.NewController(nil)
	} else {
		ctrl = gomock.NewController(t)
	}
	cl := NewMockClient(ctrl)

	cl.EXPECT().CreateAccount(gomock.Any(), gomock.Any()).Return(&external.CreateAccountResponse{
		SC:          http.StatusOK,
		EC:          "OK",
		AccountID:   uuid.NewString(),
		ReferenceID: uuid.NewString(),
	}, nil).AnyTimes()

	cl.EXPECT().DeleteTransaction(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id string, deleteType external.DeleteType, cc string) (*external.DeleteTransactionResponse, error) {
		_, ok := txIDAmounts[id]
		if !ok {
			return &external.DeleteTransactionResponse{
				SC:     http.StatusMultiStatus,
				EC:     "404",
				Status: "no payment found to delete",
			}, nil
		}

		return &external.DeleteTransactionResponse{
			SC: http.StatusOK,
		}, nil
	}).AnyTimes()

	cl.EXPECT().CreateTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, args external.CreateTransactionArgs,
	) (*external.CreateTransactionResponse, error) {
		id := uuid.NewString()
		txIDAmounts[id] = args.Amount
		if (args.Amount == "6.66" && args.Type == external.TransactionTypePush) || (args.Amount == "7.77" && args.Type == external.TransactionTypePull) {
			return &external.CreateTransactionResponse{
				SC:            http.StatusMultiStatus,
				EC:            "fail",
				TransactionID: id,
				Network:       "Mastercard",
				NetworkRC:     "111",
				Status:        string(external.TransactionStatusError),
				ApprovalCode:  "2",
			}, nil
		}
		return &external.CreateTransactionResponse{
			SC:            http.StatusOK,
			EC:            "OK",
			TransactionID: id,
			Network:       "Mastercard",
			NetworkRC:     "000",
			Status:        string(external.TransactionStatusOk),
			ApprovalCode:  "2",
		}, nil
	}).AnyTimes()

	cl.EXPECT().RetrieveTransaction(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, id string) (*external.RetrieveTransactionResponse, error) {
		amt := "10.00"
		if _, ok := txIDAmounts[id]; ok {
			amt = txIDAmounts[id]
		}
		return &external.RetrieveTransactionResponse{
			SC:          http.StatusOK,
			EC:          "OK",
			ReferenceID: uuid.NewString(),
			Network:     "Mastercard",
			NetworkRC:   "000",
			Status:      string(external.TransactionStatusOk),
			Originally:  "OK",
			Amount:      amt,
			AmountUSD:   amt,
		}, nil
	}).AnyTimes()

	cl.EXPECT().QueryCard(gomock.Any(), gomock.Any()).Return(&external.QueryCardResponse{
		SC: http.StatusOK,
		EC: "OK",
		EM: "",
		Card: external.CardResponse{
			NameFI:         "",
			Last4:          "1523",
			Bin:            "123",
			ExpirationDate: "10/26",
			Push: external.PushObject{
				Enabled:      true,
				Network:      "Mastercard",
				Type:         "Credit",
				Regulated:    false,
				Currency:     "USD",
				Country:      "US",
				Availability: "Immediate",
			},
			Pull: external.PullObject{
				Enabled:   true,
				Network:   "Mastercard",
				Type:      "Debit",
				Regulated: false,
				Currency:  "USD",
				Country:   "US",
			},
		},
		AVS: external.AVSResponse{
			CodeAVS: external.AVSResponseCodeY,
		},
	}, nil).AnyTimes()

	cl.EXPECT().Init3DS(gomock.Any(), gomock.Any()).Return(&external.Init3DSResponse{
		SC:    http.StatusOK,
		EC:    "OK",
		ID3DS: uuid.NewString(),
		JWT:   uuid.NewString(),
	}, nil).AnyTimes()

	cl.EXPECT().Lookup3DS(gomock.Any(), gomock.Any()).Return(&external.Lookup3DSResponse{
		SC:                     http.StatusOK,
		EC:                     "OK",
		Version3DS:             "1",
		Enrolled:               "YES",
		ProcessorTransactionID: uuid.NewString(),
		DsTransactionID:        uuid.NewString(),
		Status:                 string(external.TransactionStatusOk),
		ECI:                    "05|02",
	}, nil).AnyTimes()

	cl.EXPECT().Authenticate3DS(gomock.Any(), gomock.Any()).Return(&external.Authenticate3DSResponse{
		SC:         http.StatusOK,
		EC:         "OK",
		Version3DS: "1",
		Enrolled:   "true",
		Status:     string(external.TransactionStatusOk),
	}, nil).AnyTimes()

	return cl
}
