package ledger_test

import (
	"context"
	"testing"

	"gitlab.com/fynbos/pacioli/db"
	"gitlab.com/fynbos/pacioli/ledger"

	test_utils "gitlab.com/fynbos/pacioli/utils"
)

type TestContainer struct {
	b   ledger.Backends
	Ctx context.Context
}

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx

	_, db := db.MigrateTestDB(t, ctx)

	c.b = test_utils.NewBackends(t, db)
	return c, nil
}

func (c *TestContainer) Cleanup() error {
	return nil
}
