package grpc

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/openpayments"
	"gitlab.com/fynbos/backend/paymentpointers"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	backendv1 "gitlab.com/fynbos/proto/backend/v1"
)

func TestCreateContact(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &user.User{
		ID: uuid.NewString(),
	}
	_, err := c.Users().CreateNewWallet(context.Background(), user.CreateWalletArgs{
		UserID: u.ID,
		Name:   "Marko Polo",
	})
	require.NoError(t, err)

	// Create contact
	uc := &user.User{
		ID: uuid.NewString(),
	}
	contactWallet, err := c.Users().CreateNewWallet(context.Background(), user.CreateWalletArgs{
		UserID: uc.ID,
		Name:   "Alice bob",
	})
	require.NoError(t, err)
	pp, err := paymentpointers.Parse("$fynbos.me/alice")
	require.NoError(t, err)

	c.OPClient.EXPECT().GetPaymentPointer(gomock.Any(), pp.String()).Return(&openpayments.PaymentPointer{
		ID:         uuid.NewString(),
		URL:        pp.String(),
		WalletID:   contactWallet.ID,
		Alias:      "Test",
		Asset:      "USD",
		AssetScale: 2,
	}, nil).AnyTimes()

	c.ContactsClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(
		&contacts.Contact{
			ID:             uuid.NewString(),
			Name:           contactWallet.Name,
			PaymentPointer: pp,
			WalletID:       contactWallet.ID,
		},
		nil,
	).AnyTimes()

	rpc, err := client.CreateContact(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.CreateContactRequest{
		PaymentPointer: pp.ShortString(),
	})
	require.NoError(t, err)

	assert.Equal(t, contactWallet.ID, rpc.WalletId)
}

func TestListContacts(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &user.User{
		ID: uuid.NewString(),
	}
	wallet, err := c.Users().CreateNewWallet(context.Background(), user.CreateWalletArgs{
		UserID: u.ID,
		Name:   "Marko Polo",
	})
	require.NoError(t, err)

	// Create contact
	uc := &user.User{
		ID: uuid.NewString(),
	}
	contactWallet, err := c.Users().CreateNewWallet(context.Background(), user.CreateWalletArgs{
		UserID: uc.ID,
		Name:   "Alice bob",
	})
	require.NoError(t, err)
	pp, err := paymentpointers.Parse("$fynbos.me/alice")
	require.NoError(t, err)

	c.ContactsClient.EXPECT().List(gomock.Any(), wallet.ID, gomock.Any(), gomock.Any()).Return([]contacts.Contact{
		{
			ID:             uuid.NewString(),
			Name:           contactWallet.Name,
			PaymentPointer: pp,
			WalletID:       contactWallet.ID,
		},
	},
		nil,
	).AnyTimes()

	userContacts, err := client.ListContacts(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.ListContactsRequest{})
	require.NoError(t, err)

	assert.Len(t, userContacts.GetContacts(), 1)
}
