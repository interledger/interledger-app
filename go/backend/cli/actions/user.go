package actions

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/bxcodec/faker/v3"
	kratos "github.com/ory/kratos-client-go"
	"github.com/urfave/cli/v2"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/signup"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

var MakeUserFlags = []cli.Flag{
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
}

func MakeUser(b Backends) cli.ActionFunc {
	return func(cCtx *cli.Context) error {
		phone := 10_000_000 - rand.Int63n(999) // faker E164Phonenumber format is not accepted
		buf := make([]byte, 32)
		_, err := cryptorand.Read(buf)
		if err != nil {
			return err
		}
		password := base64.StdEncoding.EncodeToString(buf)
		activeState := kratos.IdentityState("active")
		ctx := context.WithValue(cCtx.Context, kratos.ContextServerIndex, 1)
		req := b.Kratos().IdentityApi.CreateIdentity(ctx).
			CreateIdentityBody(kratos.CreateIdentityBody{
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
			})
		identity, response, err := req.Execute()
		if err != nil {
			return err
		}
		if response.StatusCode != http.StatusCreated {
			return fmt.Errorf("Kratos statusCode: %d", response.StatusCode)
		}

		userID := identity.Id
		signupID, err := b.Signup().SetUserData(cCtx.Context, signup.UserDataArgs{
			ID:          userID,
			FirstName:   cCtx.String("firstName"),
			LastName:    cCtx.String("lastName"),
			Email:       cCtx.String("email"),
			CountryCode: cCtx.String("country"),
		})
		if err != nil {
			return err
		}
		err = b.Signup().SetMobileNumber(cCtx.Context, signup.MobileNumberArgs{
			ID:           signupID,
			MobileNumber: fmt.Sprintf("+2782%d", phone),
			OTP:          "123456",
		})
		if err != nil {
			return err
		}
		err = b.Signup().Complete(cCtx.Context, signupID, userID)
		if err != nil {
			return err
		}

		wallet, err := b.Users().CreateNewWallet(cCtx.Context, user.CreateWalletArgs{
			UserID: userID,
		})
		if err != nil {
			return err
		}

		log.Info(
			"Created user",
			zap.String("userID", userID),
			zap.String("walletID", wallet.ID),
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
					CountryCode: "GH",
					State:       "US-CA",
				},
				IPAddress: "10.10.10.10",
			})
			if err != nil {
				return err
			}
		}

		return nil
	}
}
