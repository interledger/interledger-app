package actions

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/pti"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/wallets"

	"github.com/bxcodec/faker/v3"
	kratos "github.com/ory/kratos-client-go"
	"github.com/urfave/cli/v2"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var MakeWalletFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "email",
		Usage: "`<email>` of the user",
		Value: faker.Email(),
	},
	&cli.StringFlag{
		Name:  "country",
		Usage: "`<alpha_2 code>` of the country",
		Value: "US",
	},
	&cli.StringFlag{
		Name:  "firstName",
		Usage: "`<firstName>` of the user",
		Value: faker.FirstName(),
	},
	&cli.StringFlag{
		Name:  "lastName",
		Usage: "`<firstName>` of the user",
		Value: faker.LastName(),
	},
	&cli.BoolFlag{
		Name:    "kyc",
		Aliases: []string{"k"},
		Usage:   "adds dummy kyc data to user",
		Value:   false,
	},
	&cli.BoolFlag{
		Name:    "linkedaccount",
		Aliases: []string{"l"},
		Usage:   "adds a send and receive enabled linked account to the wallet",
		Value:   false,
	},
}

func MakeWallet(b Backends) cli.ActionFunc {
	return func(cCtx *cli.Context) error {
		phone := 10_000_000 - rand.Int63n(999) // faker E164Phonenumber format is not accepted
		password := "fynboslocal"
		activeState := kratos.IdentityState("active")
		ctx := context.WithValue(cCtx.Context, kratos.ContextServerIndex, 1)
		identity, response, err := b.Kratos().IdentityApi.CreateIdentity(ctx).CreateIdentityBody(kratos.CreateIdentityBody{
			Credentials: &kratos.IdentityWithCredentials{
				Password: &kratos.IdentityWithCredentialsPassword{
					Config: &kratos.IdentityWithCredentialsPasswordConfig{
						Password: &password,
					},
				},
			},
			Traits: map[string]interface{}{
				"email":       cCtx.String("email"),
				"firstName":   cCtx.String("firstName"),
				"lastName":    cCtx.String("lastName"),
				"countryCode": cCtx.String("country"),
				"phone":       fmt.Sprintf("+2782%d", phone),
			},
			State: &activeState,
		}).Execute()
		if err != nil {
			return err
		}
		if response.StatusCode != http.StatusCreated {
			return fmt.Errorf("Kratos statusCode: %d", response.StatusCode)
		}

		address, err := wallets.ParseAddress(fmt.Sprintf("https://local.ilp.link/%s", cCtx.String("firstName")))
		if err != nil {
			return err
		}
		wallet, err := b.Wallets().Create(cCtx.Context, wallets.CreateArgs{
			UserID:    identity.Id,
			Name:      cCtx.String("firstName"),
			Addresses: []wallets.Address{address},
		})
		if err != nil {
			return err
		}

		log.Info(
			"Created user",
			zap.String("userID", identity.Id),
			zap.String("walletID", wallet.ID),
			zap.String("walletAddress", address.String()),
			zap.String("email", cCtx.String("email")),
			zap.String("password", password),
		)

		if cCtx.Bool("kyc") {
			log.Info("Updating user kyc details.")
			dob, err := time.Parse("2006-01-02", "2001-01-01")
			if err != nil {
				return err
			}
			_, err = b.KYC().UpdateIndividualDetails(cCtx.Context, kyc.IndividualDetails{
				WalletID:    wallet.ID,
				FirstName:   cCtx.String("firstName"),
				LastName:    cCtx.String("lastName"),
				CountryCode: cCtx.String("country"),
				Gender:      0,
				DateOfBirth: dob,
				Address: &kyc.Address{
					City:        "Santa Clara",
					ZipCode:     "95053",
					Line1:       "500 El Camino Real Santa Clara",
					CountryCode: "US",
					State:       "US-CA",
				},
				IPAddress: "10.10.10.10",
			})
			if err != nil {
				return err
			}

			err = b.KYC().SetKYCStatus(cCtx.Context, wallet.ID, kyc.StatusLevel1)
			if err != nil {
				return err
			}
		}

		if cCtx.Bool("linkedaccount") {
			a, err := b.PTI().CreateWallet(cCtx.Context, wallet.ID, currency.USD)
			if err != nil {
				return err
			}

			var la linkedaccounts.LinkedAccount
			err = a(cCtx.Context, &la)
			if err != nil {
				return err
			}

			log.Info(
				"Created balance account.",
				zap.String("id", la.ID),
				zap.Bool("canSend", la.CanSend),
				zap.Bool("canReceive", la.CanReceive),
				zap.String("provider", pti.ProviderName),
				zap.String("providerID", la.ProviderID),
				zap.String("currency", currency.USD.String()),
			)
		}

		return nil
	}
}
