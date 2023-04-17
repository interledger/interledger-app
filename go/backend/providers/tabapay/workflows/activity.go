package workflows

import (
	"context"
	"crypto/x509"
	"fmt"
	"os"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	external_client "gitlab.com/fynbos/backend/providers/tabapay/external/client"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

type Activity struct {
	b Backends
}

func NewActivity(cb InputBackends) *Activity {
	clientArgs := external_client.NewClientArgs{
		VgsProxyURL: os.Getenv("VGS_PROXY_URL"),
		ClientID:    os.Getenv("TABAPAY_CLIENT_ID"),
		BearerToken: os.Getenv("TABAPAY_BEARER_TOKEN"),
	}
	if os.Getenv("VGS_CERT_PATH") != "" {
		vgsCaCert, err := os.ReadFile(os.Getenv("VGS_CERT_PATH"))
		if err != nil {
			log.Fatal("Failed to read VGS certificate.", zap.Error(err))
		}

		caCertPool := x509.NewCertPool()
		ok := caCertPool.AppendCertsFromPEM(vgsCaCert)
		if !ok {
			log.Fatal("Failed to add VGS CA to cert pool.")
		}

		clientArgs.CaCertPool = caCertPool
	}

	externalClient, err := external_client.New(clientArgs)
	if err != nil {
		log.Fatal("Failed to create Tabapay activity.", zap.Error(err))
	}

	return &Activity{b: &backends{
		b:        cb,
		external: externalClient,
	}}
}

func (a *Activity) CreateExternalCard(ctx context.Context, args CreateExternalCardArgs) (*external.CreateAccountResponse, error) {
	owner, err := a.b.KYC().GetIndividualDetails(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}
	if owner.Address == nil {
		return nil, fmt.Errorf("%w require address for wallet.", tabapay.ErrInternal)
	}

	resp, err := a.b.External().CreateAccount(ctx, external.CreateAccountArgs{
		ReferenceID: args.LinkedAccountID[:15], // tabapay requires 1 < len(ReferenceID) < 15
		Card: external.Card{
			AccountNumber:  args.CardNumber,
			ExpirationDate: args.ExpirationDate,
		},
		Owner: external.Owner{
			Name: external.Name{
				First: owner.FirstName,
				Last:  owner.LastName,
			},
			Address: &external.Address{
				Line1:   owner.Address.Line1,
				Line2:   owner.Address.Line2,
				City:    owner.Address.City,
				State:   owner.Address.State,
				ZipCode: owner.Address.ZipCode,
				Country: owner.Address.CountryCode,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return resp, nil
}

func (a *Activity) CreateLinkedCard(ctx context.Context, args CreateLinkedCardArgs) (*linkedaccounts.LinkedAccount, error) {
	la, err := a.b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		ID:         args.ID,
		WalletID:   args.WalletID,
		Name:       args.Name,
		Mask:       args.Mask,
		Provider:   tabapay.ProviderName,
		ProviderID: args.ProviderID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return la, nil
}
