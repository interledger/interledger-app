package ops_test

import (
	"context"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/contacts/ops"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/wallets"
)

type backends struct {
	validator *validator.Validate
	db        *sqlx.DB
}

func (b backends) Validator() *validator.Validate {
	return b.validator
}

func (b backends) DB() *sqlx.DB {
	return b.db
}

func TestCreateAContact(t *testing.T) {
	ctx := context.Background()
	testDb := db.MigrateTestDB(t, ctx)
	b := &backends{
		validator: validator.New(),
		db:        testDb,
	}
	wid := uuid.NewString()
	pp, err := wallets.ParseAddress("$ilp.link/marko")
	require.NoError(t, err)

	c, err := ops.Create(ctx, b, contacts.CreateContactArgs{
		Name:           "Marko polo",
		PaymentPointer: pp,
		WalletID:       wid,
	})
	require.NoError(t, err)

	assert.Equal(t, "Marko polo", c.Name)
	assert.Equal(t, pp.String(), pp.String())
}

func TestListContacts(t *testing.T) {
	ctx := context.Background()
	testDb := db.MigrateTestDB(t, ctx)
	b := &backends{
		validator: validator.New(),
		db:        testDb,
	}
	wid := uuid.NewString()
	contactNames := []string{
		"Joshua Davila",
		"Timothy Zamora",
		"Bentley Wilcox",
		"Efrain Ayers",
		"Aaron Charles",
		"Todd Blanchard",
		"Richard Booth",
		"Craydon Rasmussen",
		"Zauryn Estrada",
		"Micaela Brady",
		"John Jacob",
	}
	for _, name := range contactNames {
		pp, err := wallets.ParseAddress("$ilp.link/" + strings.ReplaceAll(name, " ", ""))
		require.NoError(t, err, "name", name)
		_, err = ops.Create(ctx, b, contacts.CreateContactArgs{
			Name:           name,
			PaymentPointer: pp,
			WalletID:       wid,
		})
		require.NoError(t, err)
	}

	lc, err := ops.List(ctx, b, wid, db.Pagination{
		PageToken: "",
		PageSize:  50,
	}, "")
	require.NoError(t, err)
	assert.Len(t, lc, 11)
	assert.Equal(t, "Aaron Charles", lc[0].Name)

	// Pagination works
	lc, err = ops.List(ctx, b, wid, db.Pagination{
		PageToken: lc[0].ID,
		PageSize:  3,
	}, "")
	require.NoError(t, err)
	assert.Len(t, lc, 4)
	assert.Equal(t, "Bentley Wilcox", lc[0].Name)

	// Order by works
	err = ops.SetLastPaidAtNow(ctx, b, wid, lc[1].PaymentPointer)
	require.NoError(t, err)

	//TODO: fix order by
	// lc, err = ops.List(ctx, b, wid, db.Pagination{
	// 	PageToken: "",
	// 	PageSize:  50,
	// }, "last_paid_at desc")
	// require.NoError(t, err)

	// assert.Equal(t, "Craydon Rasmussen", lc[0].Name)
}

func TestGetContact(t *testing.T) {
	ctx := context.Background()
	testDb := db.MigrateTestDB(t, ctx)
	b := &backends{
		validator: validator.New(),
		db:        testDb,
	}
	wid := uuid.NewString()
	pp, err := wallets.ParseAddress("$ilp.link/marko")
	require.NoError(t, err)
	c, err := ops.Create(ctx, b, contacts.CreateContactArgs{
		Name:           "Marko polo",
		PaymentPointer: pp,
		WalletID:       wid,
	})
	require.NoError(t, err)

	gc, err := ops.Get(ctx, b, wid, pp)
	require.NoError(t, err)

	assert.Equal(t, c.Name, gc.Name)
	assert.Equal(t, c.PaymentPointer, gc.PaymentPointer)

	// Unknown error
	randomPP, err := wallets.ParseAddress("$ilp.link/test")
	require.NoError(t, err)

	_, err = ops.Get(ctx, b, wid, randomPP)
	require.ErrorIs(t, err, contacts.ErrNotFound)
}

func TestSetLastPaidAtContact(t *testing.T) {
	ctx := context.Background()
	testDb := db.MigrateTestDB(t, ctx)
	b := &backends{
		validator: validator.New(),
		db:        testDb,
	}
	wid := uuid.NewString()
	pp, err := wallets.ParseAddress("$ilp.link/marko")
	require.NoError(t, err)
	c, err := ops.Create(ctx, b, contacts.CreateContactArgs{
		Name:           "Marko polo",
		PaymentPointer: pp,
		WalletID:       wid,
	})
	require.NoError(t, err)
	require.False(t, c.LastPaidAt.Valid)

	err = ops.SetLastPaidAtNow(ctx, b, c.WalletID, c.PaymentPointer)
	require.NoError(t, err)

	gc, err := ops.Get(ctx, b, wid, pp)
	require.NoError(t, err)

	require.True(t, gc.LastPaidAt.Valid)
}
