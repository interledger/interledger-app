package ops

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/mx"
	"gitlab.com/fynbos/backend/providers/mx/external"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.uber.org/zap"
)

func GetWidget(ctx context.Context, b Backends, walletID string) (string, error) {
	externalUsers, err := b.External().ListUsersByID(ctx, walletID)
	if err != nil {
		return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	var user *external.User
	for _, usr := range externalUsers.Users {
		if usr.ID == walletID {
			user = &usr
			break
		}
	}
	if user == nil {
		user, err = b.External().CreateUser(ctx, walletID)
		if err != nil {
			return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
		}
	}

	widget, err := b.External().GetWidgetURL(ctx, external.GetWidgetURLArgs{
		UserGuid:            user.Guid,
		IncludeTransactions: false,
		IncludeIdentity:     true,
		Mode:                "verification",
		WidgetType:          "connect_widget",
	})
	if err != nil {
		return "", fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return widget.URL, nil
}

func CreateBankAccounts(ctx context.Context, b Backends, args mx.CreateBankAccountsArgs) ([]linkedaccounts.LinkedAccount, error) {
	accountOwnersResponse, err := b.External().ListAccountOwnersByMember(ctx, args.UserGuid, args.MemberGuid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	walletOwner, err := b.KYC().GetIndividualDetails(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	existingAccounts, err := b.LinkedAccounts().ListMXBankAccounts(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	type createAccountInfo struct {
		Guid  string
		State linkedaccounts.State
	}
	var accountsToCreate []createAccountInfo
	for _, accountOwner := range accountOwnersResponse.AccountOwners {
		var skip bool
		for _, existingAccount := range existingAccounts {
			if accountOwner.AccountGuid == existingAccount.ProviderID {
				skip = true
				break
			}
		}
		if skip {
			continue
		}

		state := linkedaccounts.Verified
		if !isAccountOwner(walletOwner, accountOwner) {
			state = linkedaccounts.OwnershipReviewRequired
		}
		accountsToCreate = append(accountsToCreate, createAccountInfo{
			Guid:  accountOwner.AccountGuid,
			State: state,
		})
	}

	accountsResponse, err := b.External().ListAccountsByMember(ctx, args.UserGuid, args.MemberGuid)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	var createLinkedAccounts []linkedaccounts.CreateArgs
	for _, accountToCreate := range accountsToCreate {
		for _, account := range accountsResponse.Accounts {
			// only add checking or savings
			if account.Type != mx.TypeChecking && account.Type != mx.TypeSavings {
				continue
			}

			if account.GUID == accountToCreate.Guid {
				mask := account.AccountNumber
				if len(mask) > 4 {
					mask = mask[len(mask)-4:]
				}
				createLinkedAccounts = append(createLinkedAccounts, linkedaccounts.CreateArgs{
					WalletID:   args.WalletID,
					Name:       account.Name,
					Nickname:   account.Nickname,
					Mask:       mask,
					Provider:   mx.ProviderName,
					ProviderID: account.GUID,
					Type:       mx.TypeBankAccount,
					CanSend:    true,
					CanReceive: true,
					State:      accountToCreate.State,
				})
			}
		}
	}

	las, err := b.LinkedAccounts().CreateBatch(ctx, createLinkedAccounts)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	var reviews []linkedaccounts.CreateReviewArgs
	for _, la := range las {
		if la.State != linkedaccounts.Verified {
			reviews = append(reviews, linkedaccounts.CreateReviewArgs{
				LinkedAccountID: la.ID,
				State:           la.State,
			})
		}
	}

	_, err = b.LinkedAccounts().CreateReviews(ctx, reviews)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return las, nil
}

func isAccountOwner(walletOwner *kyc.IndividualDetails, accountOwner external.AccountOwners) bool {
	if !env.IsProd() {
		log.Info(
			"checking mx account ownership",
			zap.String("mx account guid", accountOwner.AccountGuid),
			zap.String("mx owner name", accountOwner.OwnerName),
			zap.String("fynbos wallet owner name", fmt.Sprintf("%s %s", walletOwner.FirstName, walletOwner.LastName)),
		)
		return true
	}

	return strings.EqualFold(fmt.Sprintf("%s %s", walletOwner.FirstName, walletOwner.LastName), accountOwner.OwnerName)
}

func GetAccount(ctx context.Context, b Backends, walletID, guid string) (*mx.Account, error) {
	externalUsers, err := b.External().ListUsersByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	var user *external.User
	for _, u := range externalUsers.Users {
		if u.ID == walletID {
			user = &u
			break
		}
	}
	if user == nil {
		return nil, mx.ErrNotFound
	}

	externalAccount, err := b.External().ReadUsersAccount(ctx, user.Guid, guid)
	if errors.Is(err, external.ErrNotFound) {
		return nil, mx.ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	updatedAt, err := time.Parse(time.RFC3339, externalAccount.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("%w %s", mx.ErrInternal, err)
	}

	return &mx.Account{
		Guid:             externalAccount.GUID,
		MemberGuid:       externalAccount.MemberGUID,
		UserGuid:         externalAccount.UserGUID,
		Name:             externalAccount.Name,
		Nickname:         externalAccount.Nickname,
		AccountNumber:    externalAccount.AccountNumber,
		RoutingNumber:    externalAccount.RoutingNumber,
		InstitutionCode:  externalAccount.InstitutionCode,
		IsHidden:         externalAccount.IsHidden,
		Type:             externalAccount.Type,
		UpdatedAt:        updatedAt,
		Balance:          currency.FromFloat64(externalAccount.Balance, currency.Currency(externalAccount.CurrencyCode)),
		AvailableBalance: currency.FromFloat64(externalAccount.AvailableBalance, currency.Currency(externalAccount.CurrencyCode)),
	}, nil
}
