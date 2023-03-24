package ops

import (
	"context"
	"fmt"

	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
)

func CreateCard(ctx context.Context, b Backends, args tabapay.CreateCardArgs) (*linkedaccounts.LinkedAccount, error) {
	owner, err := b.KYC().GetIndividualDetails(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}
	if owner.Address == nil {
		return nil, fmt.Errorf("%w require address for wallet.", tabapay.ErrInternal)
	}

	resp, err := b.External().CreateAccount(ctx, external.CreateAccountArgs{
		ReferenceID: args.ReferenceID,
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

	la, err := b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		ID:         args.ReferenceID,
		WalletID:   args.WalletID,
		Name:       args.Name,
		Provider:   tabapay.ProviderName,
		ProviderID: resp.AccountID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return la, nil
}
