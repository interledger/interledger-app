package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gitlab.com/fynbos/backend/accounts"
	"gitlab.com/fynbos/backend/providers/rafiki"
	_user "gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestGetQuote(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	user := &_user.User{
		ID: uuid.NewString(),
	}

	receiveAmount := uint64(99)
	sendAmount := uint64(100)
	assetScale := uint8(2)
	currencyCode := "USD"
	identifierID := uuid.NewString()
	accountID := uuid.NewString()
	receiverPaymentPointer := "$ilp.test/receiver"
	c.AccountService.EXPECT().GetByIdentityID(gomock.Any(), user.ID).Return(
		&accounts.Account{
			ID:         accountID,
			IdentityID: user.ID,
		},
		nil,
	).Times(1)
	c.RafikiProvider.EXPECT().GetIdentifierByAccountAndCurrency(
		gomock.Any(),
		accountID,
		currencyCode,
	).Return(
		&rafiki.Identifier{
			ID:         identifierID,
			AccountID:  accountID,
			AssetCode:  currencyCode,
			AssetScale: assetScale,
		},
		nil,
	).Times(1)
	expiresAt := time.Now().Format(time.RFC3339)
	c.RafikiProvider.EXPECT().CreateQuote(gomock.Any(),
		&rafiki.CreateQuoteArgs{
			IdentifierID:           identifierID,
			ReceiverPaymentPointer: receiverPaymentPointer,
			SendAssetCode:          currencyCode,
			SendAssetScale:         assetScale,
			SendAmount:             sendAmount,
		},
	).Return(
		&rafiki.Quote{
			ID:                        uuid.NewString(),
			IdentifierID:              identifierID,
			ExpiresAt:                 expiresAt,
			Receiver:                  receiverPaymentPointer,
			SendAssetCode:             currencyCode,
			SendAssetScale:            assetScale,
			SendAmount:                sendAmount,
			ReceiveAmount:             receiveAmount,
			ReceiveAssetCode:          currencyCode,
			ReceiveAssetScale:         assetScale,
			MinExchangeRate:           1,
			LowEstimatedExchangeRate:  1,
			HighEstimatedExchangeRate: 1,
		},
		nil,
	).Times(1)

	quote, err := client.GetQuote(
		user_mock.ActingAsContext(t, context.Background(), user),
		&backendv1.GetQuoteRequest{
			SendAmount:             sendAmount,
			SendCurrencyCode:       currencyCode,
			ReceiverPaymentPointer: receiverPaymentPointer,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, sendAmount, quote.SendAmount)
	assert.Equal(t, currencyCode, quote.SendCurrencyCode)
	assert.Equal(t, receiveAmount, quote.ReceiveAmount)
	assert.Equal(t, currencyCode, quote.ReceiveCurrencyCode)
	assert.Equal(t, expiresAt, quote.ExpiresAt)
}
