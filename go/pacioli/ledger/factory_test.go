package ledger_test

import (
	"context"
	"fmt"
	"testing"

	"gitlab.com/fynbos/pacioli/ledger"

	tigerbeetle_go "github.com/coilhq/tigerbeetle-go"
	"github.com/jmoiron/sqlx"
	test_utils "gitlab.com/fynbos/pacioli/utils"
)

type TestContainer struct {
	TbClient tigerbeetle_go.Client
	Ctx      context.Context
	Tb       *test_utils.TigerBeetleContainer
	Db       *sqlx.DB
	Ls       ledger.Service
}

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx
	containerNetwork := "pacioli-test"

	_, c.Db = test_utils.MigrateCockroachDB(t, ctx)

	tb, err := test_utils.SetupTigerBeetle(ctx, 0, containerNetwork)
	if err != nil {
		return nil, err
	}
	c.Tb = tb

	tbClient, err := tigerbeetle_go.NewClient(0, []string{tb.URI}, 1000)
	c.TbClient = tbClient

	if err != nil {
		fmt.Println()
		fmt.Println(tb.URI)
		fmt.Println(err)
		fmt.Println()
		return nil, err
	}

	ls, err := ledger.NewService(&ledger.ServiceArgs{
		Db: c.Db,
		Tb: tbClient,
	})
	if err != nil {
		return nil, err
	}
	c.Ls = ls

	return c, nil
}

func (c *TestContainer) Cleanup() error {
	c.TbClient.Close()

	err := c.Tb.Terminate(c.Ctx)
	if err != nil {
		return err
	}

	return nil
}
