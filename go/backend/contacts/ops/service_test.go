package ops_test

import (
	"context"
	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/contacts"
	"gitlab.com/fynbos/backend/contacts/ops"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/fynbos/backend/db"
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

	c, err := ops.Create(ctx, b, contacts.CreateContactArgs{
		Name:           "Marko polo",
		PaymentPointer: "$fynbos.me/marko",
		WalletID:       wid,
	})
	require.NoError(t, err)

	assert.Equal(t, "Marko polo", c.Name)
	assert.Equal(t, "$fynbos.me/marko", c.PaymentPointer)
}

func TestListContacts(t *testing.T) {
	ctx := context.Background()
	testDb := db.MigrateTestDB(t, ctx)
	b := &backends{
		validator: validator.New(),
		db:        testDb,
	}
	wid := uuid.NewString()
	c, err := ops.Create(ctx, b, contacts.CreateContactArgs{
		Name:           "Marko polo",
		PaymentPointer: "$fynbos.me/marko",
		WalletID:       wid,
	})
	require.NoError(t, err)

	lc, err := ops.List(ctx, b, wid, db.Pagination{
		PageToken: "",
		PageSize:  50,
	})
	require.NoError(t, err)

	assert.Len(t, lc, 1)
	assert.Equal(t, c.Name, lc[0].Name)
	assert.Equal(t, c.PaymentPointer, lc[0].PaymentPointer)

	// Pagination works
	lc, err = ops.List(ctx, b, wid, db.Pagination{
		PageToken: c.ID,
		PageSize:  50,
	})
	require.NoError(t, err)

	assert.Len(t, lc, 0)
}

func TestGetContact(t *testing.T) {
	ctx := context.Background()
	testDb := db.MigrateTestDB(t, ctx)
	b := &backends{
		validator: validator.New(),
		db:        testDb,
	}
	wid := uuid.NewString()
	c, err := ops.Create(ctx, b, contacts.CreateContactArgs{
		Name:           "Marko polo",
		PaymentPointer: "$fynbos.me/marko",
		WalletID:       wid,
	})
	require.NoError(t, err)

	gc, err := ops.Get(ctx, b, wid, c.PaymentPointer)
	require.NoError(t, err)

	assert.Equal(t, c.Name, gc.Name)
	assert.Equal(t, c.PaymentPointer, gc.PaymentPointer)

	// Unknown error
	_, err = ops.Get(ctx, b, wid, "random")
	require.ErrorIs(t, err, contacts.ErrNotFound)
}
