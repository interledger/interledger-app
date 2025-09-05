package grpc

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/user"
	user_mock "gitlab.com/fynbos/backend/user/client/mock"
	"gitlab.com/fynbos/backend/wallets"
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
	w := wallets.Wallet{
		ID:   uuid.NewString(),
		Name: "testing",
	}
	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&w, nil).AnyTimes()

	// Create contact
	wa, err := wallets.ParseAddress("$ilp.link/alice")
	require.NoError(t, err)
	cw := wallets.Wallet{
		ID:        uuid.NewString(),
		Name:      "testing",
		Addresses: []wallets.Address{wa},
	}
	c.walletImpl.EXPECT().Get(gomock.Any(), cw.ID).Return(&cw, nil).AnyTimes()

	c.walletImpl.EXPECT().GetFromAddress(gomock.Any(), gomock.Any()).Return(&cw, nil).AnyTimes()

	c.ContactsClient.EXPECT().Create(gomock.Any(), gomock.Any()).Return(
		&contacts.Contact{
			ID:             uuid.NewString(),
			Name:           cw.Name,
			PaymentPointer: wa,
			WalletID:       cw.ID,
		},
		nil,
	).AnyTimes()

	rpc, err := client.CreateContact(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.CreateContactRequest{
		PaymentPointer: wa.ShortString(),
	})
	require.NoError(t, err)

	assert.Equal(t, cw.ID, rpc.WalletId)
}

func TestListContacts(t *testing.T) {
	t.Parallel()
	ctrl := gomock.NewController(t)
	c := NewTestContainer(t, ctrl)
	_, _, client := startTestServer(t, c)
	u := &user.User{
		ID: uuid.NewString(),
	}
	w := wallets.Wallet{
		ID:   uuid.NewString(),
		Name: "testing",
	}
	c.walletImpl.EXPECT().List(gomock.Any(), u.ID).Return([]wallets.Wallet{w}, nil).AnyTimes()
	c.walletImpl.EXPECT().ForContext(gomock.Any()).Return(&w, nil).AnyTimes()

	// Create contact
	cw := wallets.Wallet{
		ID:   uuid.NewString(),
		Name: "testing",
	}
	c.walletImpl.EXPECT().Get(gomock.Any(), cw.ID).Return(&cw, nil).AnyTimes()

	pp, err := wallets.ParseAddress("$ilp.link/alice")
	require.NoError(t, err)

	c.ContactsClient.EXPECT().List(gomock.Any(), w.ID, gomock.Any(), gomock.Any()).Return([]contacts.Contact{
		{
			ID:             uuid.NewString(),
			Name:           cw.Name,
			PaymentPointer: pp,
			WalletID:       cw.ID,
		},
	},
		nil,
	).AnyTimes()

	userContacts, err := client.ListContacts(user_mock.ActingAsContext(t, context.Background(), u), &backendv1.ListContactsRequest{})
	require.NoError(t, err)

	assert.Len(t, userContacts.GetContacts(), 1)
}
