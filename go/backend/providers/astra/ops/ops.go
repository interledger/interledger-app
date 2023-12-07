package ops

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/fynbos/backend/providers/astra"
	"gitlab.com/fynbos/backend/providers/astra/external"
)

func CreateIntent(ctx context.Context, b Backends, walletID string) error {
	id, err := b.KYC().GetIndividualDetails(ctx, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", astra.ErrInternal, err)
	}
	if id.Address == nil {
		return fmt.Errorf("%w incomplete KYC for astra, missing address", astra.ErrNotFound)
	}

	idNums, err := b.KYC().GetPersonaIDNumbers(ctx, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	ul, err := b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	if len(ul) <= 0 {
		return fmt.Errorf("%w no users found for wallet", astra.ErrNotFound)
	}

	u := ul[0]

	args := external.CreateIntentReq{
		Email:          u.Email,
		Phone:          u.PhoneNumber,
		FirstName:      id.FirstName,
		LastName:       id.LastName,
		Address1:       id.Address.Line1,
		Address2:       id.Address.Line2,
		City:           id.Address.City,
		State:          id.Address.State,
		PostalCode:     id.Address.ZipCode,
		DateOfBirth:    id.DateOfBirth.Format(time.DateOnly),
		SocialSecurity: idNums.SocialSecurity,
		IPAddress:      id.IPAddress,
	}

	intentID, err := b.External().CreateIntent(ctx, args)
	if err != nil {
		return err
	}

	_, err = b.DB().ExecContext(ctx, "INSERT INTO astra_user_intents (wallet_id, intent_id, status, user_id) VALUES ($1, $2, 'unknown', '')", walletID, intentID)
	if err != nil {
		return fmt.Errorf("%w %s", astra.ErrInternal, err)
	}

	return nil
}
