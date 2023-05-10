package workflows

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/basistheory"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	external_client "gitlab.com/fynbos/backend/providers/tabapay/external/client"
	"gitlab.com/fynbos/log"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
)

type Activity struct {
	b Backends
}

func NewActivity(cb InputBackends) *Activity {
	clientArgs := external_client.NewClientArgs{
		BasisTheoryProxyApiKey: os.Getenv("BASISTHEORY_API_KEY"),
		ClientID:               os.Getenv("TABAPAY_CLIENT_ID"),
		BearerToken:            os.Getenv("TABAPAY_BEARER_TOKEN"),
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

	ctry, err := country.Country(owner.Address.CountryCode).Numeric()
	if err != nil {
		return nil, fmt.Errorf("%w invalid country=%s", tabapay.ErrInternal, owner.Address.CountryCode)
	}
	stateParts := strings.Split(owner.Address.State, "-")
	state := stateParts[0]
	if len(stateParts) == 2 {
		state = stateParts[1]
	}

	referenceID := tabapay.NewReferenceID()
	resp, err := a.b.External().CreateAccount(ctx, external.CreateAccountArgs{
		ReferenceID: referenceID,
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
				State:   state,
				ZipCode: owner.Address.ZipCode,
				Country: ctry,
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
		Nickname:   args.Name,
		Type:       tabapay.TypeCard,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return la, nil
}

func (a *Activity) MarkCardNotDeleted(ctx context.Context, id string) (*linkedaccounts.LinkedAccount, error) {
	la, err := a.b.LinkedAccounts().MarkNotDeleted(ctx, id)
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "NotFound", err)
	}
	if err != nil {
		return nil, err
	}

	return la, nil
}

func (a *Activity) CreateBasisTheoryCard(ctx context.Context, walletID, tokenID string) (*basistheory.Card, error) {
	card, err := a.b.BasisTheory().CreateCard(ctx, tokenID, walletID)
	if err != nil {
		return nil, err
	}

	return card, nil
}
