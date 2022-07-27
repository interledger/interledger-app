package ledger_test

import (
	"context"
	"fmt"
	"testing"

	"gitlab.com/fynbos/pacioli/ledger"

	tigerbeetle_go "github.com/coilhq/tigerbeetle-go"
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
	containerNetwork := "pacioli-test"

	_, db := test_utils.MigrateCockroachDB(t, ctx)

	tb, err := test_utils.SetupTigerBeetle(ctx, 0, containerNetwork)
	if err != nil {
		return nil, err
	}
	c.Tb = tb

	tbClient, err := tigerbeetle_go.NewClient(0, []string{tb.URI}, 1000)
	if err != nil {
		fmt.Println()
		fmt.Println(tb.URI)
		fmt.Println(err)
		fmt.Println()
		return nil, err
	}

	c.b = test_utils.NewBackends(t, db, tbClient)
	return c, nil
}

func (c *TestContainer) Cleanup() error {
	c.b.TigerBeetle().Close()

	err := c.Tb.Terminate(c.Ctx)
	if err != nil {
		return err
	}

	return nil
}
