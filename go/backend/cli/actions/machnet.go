package actions

import (
	"errors"

	"github.com/urfave/cli/v2"
	"gitlab.com/fynbos/backend/providers/machnet"
	machnet_external "gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var MakeMachnetSendUserFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "walletID",
		Usage:    "`walletID` for which to create send user",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "userID",
		Usage:    "Fynbos user",
		Required: true,
	},
}

func MakeMachnetSendUser(b Backends) cli.ActionFunc {
	return func(cCtx *cli.Context) error {
		ctx := cCtx.Context
		walletID := cCtx.String("walletID")
		_, err := b.Users().GetWallet(ctx, cCtx.String("userID"), walletID)
		if err != nil {
			return err
		}
		users, err := b.Users().ListUsers(ctx, walletID)
		if err != nil {
			return err
		}
		if len(users) < 1 {
			return errors.New("No users found for wallet.")
		}

		kyc, err := b.KYC().GetIndividualDetails(ctx, cCtx.String("walletID"))
		if err != nil {
			return err
		}

		externalUser, err := b.MachnetExternal().RegisterUser(ctx, machnet_external.User{
			FirstName:    kyc.FirstName,
			LastName:     kyc.LastName,
			Email:        users[0].Email,
			Country:      kyc.CountryCode,
			City:         "Santa Clara",
			Type:         machnet_external.TypeSendUser,
			AddressLine1: "500 8 El Camino Real Santa Clara",
			DateOfBirth:  "2000-01-01",
			Gender:       "female",
			State:        "CA",
			Business:     false,
			BusinessType: "LLC",
			MobilePhone:  "9879879870",
			IPAddress:    "73.85.79.9",
			Zipcode:      "95053",
		})
		if err != nil {
			return err
		}

		sendUser, err := b.Machnet().CreateUser(ctx, machnet.CreateArgs{
			WalletID:   walletID,
			ExternalID: externalUser.ID,
		})
		if err != nil {
			return err
		}

		log.Info("Created machnet send user", zap.String("external user id:", sendUser.ID))

		return nil
	}
}

var MakeReceiveBankAccountFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "walletID",
		Usage:    "`walletID` for which to create receive bank account",
		Required: true,
	},
	&cli.StringFlag{
		Name:  "name",
		Usage: "`name` of the receive bank account",
		Value: "test",
	},
}

func MakeReceiveBankAccount(b Backends) cli.ActionFunc {
	return func(cCtx *cli.Context) error {
		ctx := cCtx.Context

		la, err := b.Machnet().CreateReceiveBankAccount(ctx, machnet.CreateReceiveBankAccountArgs{
			WalletID:      cCtx.String("walletID"),
			AccountNumber: "316497",
			AccountType:   machnet.BankAccountTypeCheque,
			BankID:        37,
			BranchID:      37,
			Name:          cCtx.String("name"),
		})
		if err != nil {
			return err
		}
		log.Info("Created linked account", zap.String("linked account id:", la.ID))

		return nil
	}
}
