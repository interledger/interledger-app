package workflows

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
)

type Activity struct {
	b Backends
}

func NewActivity(b Backends) *Activity {
	return &Activity{b: b}
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
		ReferenceID: args.LinkedAccountID,
		Card: external.Card{
			AccountNumber:  args.CardNumber,
			ExpirationDate: args.ExpirationDate,
			SecurityCode:   args.CVV,
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
