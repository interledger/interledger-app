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
	Tb  *test_utils.TigerBeetleContainer
}

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx

	_, db := db.MigrateTestDB(t, ctx)

	// tbClient, err := tigerbeetle_go.NewClient(0, []string{"0.0.0.0:3000"}, 1000)
	// if err != nil {
	// 	fmt.Println()
	// 	fmt.Println("0.0.0.0:3000")
	// 	fmt.Println(err)
	// 	fmt.Println()
	// 	return nil, err
	// }

	c.b = test_utils.NewBackends(t, db, nil)
	return c, nil
}

func (c *TestContainer) Cleanup() error {
	// c.b.TigerBeetle().Close()

	return nil
}
