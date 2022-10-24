package actions

import (
	"fmt"
	"math/rand"
	"net/http"

	"github.com/bxcodec/faker/v3"
	kratos "github.com/ory/kratos-client-go"
	"github.com/urfave/cli/v2"
	"gitlab.com/fynbos/backend/signup"
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
}

func MakeUser(b Backends) cli.ActionFunc {
	return func(cCtx *cli.Context) error {
		phone := 10_000_000 - rand.Int63n(999) // faker E164Phonenumber format is not accepted
		req := b.Kratos().V0alpha2Api.AdminCreateIdentity(cCtx.Context).
			AdminCreateIdentityBody(kratos.AdminCreateIdentityBody{
				Traits: map[string]interface{}{
					"email":       cCtx.String("email"),
					"firstName":   cCtx.String("firstName"),
					"lastName":    cCtx.String("lastName"),
					"countryCode": cCtx.String("country"),
					"phone":       fmt.Sprintf("+2782%d", phone),
				},
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

		_, err = b.User().CreateNewWallet(cCtx.Context, userID, "default")
		if err != nil {
			return err
		}

		log.Info("Created user", zap.String("userID", userID))

		return nil
	}
}
