package actions

import (
	"github.com/urfave/cli/v2"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var MakeMachnetTransactionFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "fromLinkedAccountID",
		Aliases:  []string{"from"},
		Usage:    "`LinkedAccountID` of the payer",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "toLinkedAccountID",
		Aliases:  []string{"to"},
		Usage:    "`LinkedAccountID` of the payee",
		Required: true,
	},
	&cli.StringFlag{
		Name:    "currency",
		Aliases: []string{"c"},
		Usage:   "`LinkedAccountID` of the payee. Default=USD",
		Value:   "USD",
	},
	&cli.Float64Flag{
		Name:     "amount",
		Usage:    "`float64` amount to send",
		Required: true,
	},
}

func MakeMachnetTransaction(b Backends) cli.ActionFunc {
	return func(cCtx *cli.Context) error {
		ctx := cCtx.Context
		from, err := b.LinkedAccounts().Get(ctx, cCtx.String("fromLinkedAccountID"))
		if err != nil {
			return err
		}

		to, err := b.LinkedAccounts().Get(ctx, cCtx.String("toLinkedAccountID"))
		if err != nil {
			return err
		}

		await, err := b.Machnet().CreateTransaction(ctx, machnet.CreateTransactionArgs{
			FromLinkedAccountID: from.ID,
			ToLinkedAccountID:   to.ID,
			Amount:              currency.FromFloat64(cCtx.Float64("amount"), currency.ParseCurrency(cCtx.String("currency"))),
		})
		if err != nil {
			return err
		}

		var trxID string
		err = await(ctx, trxID)
		if err != nil {
			return err
		}

		log.Info("Created transaction", zap.String("transactionID", trxID))
		return nil
	}
}
