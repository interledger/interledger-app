package webhook_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/providers/tabapay"
	tabapay_mock "gitlab.com/fynbos/backend/providers/tabapay/client/mock"
	"gitlab.com/fynbos/backend/providers/verygoodsecurity"
	verygoodsecurity_mock "gitlab.com/fynbos/backend/providers/verygoodsecurity/client/mock"
	"gitlab.com/fynbos/backend/providers/verygoodsecurity/webhook"
	"go.uber.org/zap"

	"github.com/golang/mock/gomock"
)

type TestContainer struct {
	Logger        *zap.Logger
	vgs           *verygoodsecurity_mock.MockClient
	ValidatorImpl *validator.Validate
	tabapay       *tabapay_mock.MockClient
}

func (t TestContainer) VGS() verygoodsecurity.Client {
	return t.vgs
}

func (t TestContainer) Tabapay() tabapay.Client {
	return t.tabapay
}

func NewTestContainer(t *testing.T, ctrl *gomock.Controller) (*TestContainer, error) {
	c := &TestContainer{ValidatorImpl: validator.New()}

	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	c.Logger = logger

	c.vgs = verygoodsecurity_mock.NewMockClient(ctrl)
	c.tabapay = tabapay_mock.NewMockClient(ctrl)

	return c, nil
}

func TestNewHandleInboundCard(s *testing.T) {
	s.Parallel()
	ctrl := gomock.NewController(s)
	c, err := NewTestContainer(s, ctrl)
	if err != nil {
		s.Fatal(err)
	}

	walletID, cardID := uuid.NewString(), uuid.NewString()
	c.vgs.EXPECT().CreateCard(gomock.Any(), verygoodsecurity.Card{
		WalletID: walletID,
		Token:    "4111112436781111",
		Expiry:   "some_token_2937648273",
		CVV:      "some_token_7281687254",
		Last4:    "1111",
		Type:     "visa",
	}).Return(&verygoodsecurity.Card{
		ID:        cardID,
		WalletID:  walletID,
		Token:     "4111112436781111",
		Expiry:    "some_token_2937648273",
		CVV:       "some_token_7281687254",
		Last4:     "1111",
		Type:      "visa",
		CreatedAt: "",
		UpdatedAt: "",
	}, nil).AnyTimes()

	c.tabapay.EXPECT().CreateCard(gomock.Any(), tabapay.CreateCardArgs{
		IdempotencyKey: cardID,
		WalletID:       walletID,
		CardNumber:     "4111112436781111",
		ExpirationDate: "some_token_2937648273",
		CVV:            "some_token_7281687254",
		Name:           "visa 1111",
	}).Return(func(ctx context.Context, result interface{}) error { return nil }, nil).AnyTimes()

	s.Run("OPTIONS returns OK", func(t *testing.T) {
		optionsReq, err := http.NewRequest("OPTIONS", "/webhooks/verygoodsecurity/card", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(webhook.NewHandleInboundCard(c))

		handler.ServeHTTP(rr, optionsReq)

		assert.Equal(t, rr.Code, http.StatusOK)
	})

	s.Run("POST Returns 400 on invalid body", func(t *testing.T) {
		cardFromVGS := url.Values{}
		cardFromVGS.Set("something else", "foo")
		postReq, err := http.NewRequest("POST", "/webhooks/verygoodsecurity/card", strings.NewReader(cardFromVGS.Encode()))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(webhook.NewHandleInboundCard(c))

		handler.ServeHTTP(rr, postReq)

		assert.Equal(t, rr.Code, http.StatusBadRequest)
		assert.Equal(t, rr.Body.String(), "failed to parse payload\n")
	})

	s.Run("POST Successfully stores a new card", func(t *testing.T) {
		jsonBody := []byte(fmt.Sprintf(`{"card-number": "4111112436781111", "exp-date": "some_token_2937648273", "card-security-code": "some_token_7281687254","walletId": "%s","last4": "1111","cardType": "visa"}`, walletID))
		postReq, err := http.NewRequest("POST", "/webhooks/verygoodsecurity/card", bytes.NewReader(jsonBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(webhook.NewHandleInboundCard(c))

		handler.ServeHTTP(rr, postReq)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, rr.Body.String(), "")
	})
}
