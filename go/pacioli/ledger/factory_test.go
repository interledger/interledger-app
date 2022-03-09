package ledger

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	test_utils "gitlab.com/fynbos/pacioli/utils"
	"gitlab.com/fynbos/tigerbeetle_go"
)

type TestContainer struct {
	TbClient tigerbeetle_go.Client
	Ctx      context.Context
	Crdb     *test_utils.CockroachDBContainer
	Tb       *test_utils.TigerBeetleContainer
	Db       *sqlx.DB
	Ls       Service
}

func NewTestContainer(ctx context.Context, t *testing.T) (*TestContainer, error) {
	c := &TestContainer{}
	c.Ctx = ctx

	crdb, err := test_utils.SetupTestCockroachDB(ctx)
	if err != nil {
		return nil, err
	}
	c.Crdb = crdb

	db, err := sqlx.Connect("postgres", crdb.URI)
	if err != nil {
		return nil, err
	}
	c.Db = db

	tb, err := test_utils.SetupTigerBeetle(ctx, 0)
	if err != nil {
		return nil, err
	}
	c.Tb = tb

	tbClient, err := tigerbeetle_go.NewClient(0, []string{tb.URI})
	if err != nil {
		fmt.Println()
		fmt.Println(tb.URI)
		fmt.Println(err)
		fmt.Println()
		return nil, err
	}
	// drive the TB client.
	go func() {
		tick := time.Tick(20 * time.Millisecond)
		for range tick {
			tbClient.Tick()
		}
	}()

	ls, err := NewService(&ServiceArgs{
		Db: db,
		Tb: tbClient,
	})
	if err != nil {
		return nil, err
	}
	c.Ls = ls

	return c, nil
}

func (c *TestContainer) Cleanup() error {
	err := c.Db.Close()
	if err != nil {
		return err
	}

	err = c.Crdb.Container.Terminate(c.Ctx)
	if err != nil {
		return err
	}

	err = c.Tb.Terminate(c.Ctx)
	if err != nil {
		return err
	}

	return nil
}
