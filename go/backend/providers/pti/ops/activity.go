package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/external"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/temporal"
)

type Activity struct {
	b        Backends
	external external.Client
}

func NewActivity(b Backends) *Activity {
	external := external.New(external.ClientArgs{
		Transport: &http.Client{
			Transport: otelhttp.NewTransport(
				httplogger.NewTransport(http.DefaultTransport, b, nil),
			),
		},
	})

	return &Activity{
		b:        b,
		external: external,
	}
}

func (a *Activity) GetPtiUser(ctx context.Context, walletID string) (*pti.User, error) {
	var user pti.User
	err := a.b.DB().GetContext(ctx, &user, fmt.Sprintf("SELECT %s from pti_users where wallet_id=$1;", userFields), walletID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (a *Activity) CreatePtiUser(ctx context.Context, walletID string) (string, error) {
	kycData, err := a.b.KYC().GetIndividualDetails(ctx, walletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return "", temporal.NewNonRetryableApplicationError("No KYC data for wallet.", "ErrNotFound", err)
	}

	var addresses []external.Address
	if kycData.Address != nil {
		addresses = append(addresses, external.Address{
			Street:    kycData.Address.Line1,
			City:      kycData.Address.City,
			StateCode: kycData.Address.State,
			Country:   kycData.Address.CountryCode,
		})
	}

	var emails []external.Email
	var phones []external.Phone
	usrs, err := a.b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return "", err
	}

	for i, usr := range usrs {
		emails = append(emails, external.Email{
			Address: usr.Email,
			Default: i == 0,
		})

		phones = append(phones, external.Phone{
			Number:  usr.PhoneNumber,
			Type:    "MOBILE",
			Default: i == 0,
		})
	}

	return a.external.CreateUser(ctx, external.CreateUserArgs{
		ID:          uuid.NewString(),
		Type:        "Person",
		DateOfBirth: kycData.DateOfBirth.Format("2006-01-02"),
		Name: external.Name{
			First: kycData.FirstName,
			Last:  kycData.LastName,
		},
		Emails:    emails,
		Phones:    phones,
		Addresses: addresses,
	})
}

func (a *Activity) SavePtiUser(ctx context.Context, externalUserID, walletID string) (*pti.User, error) {
	var user pti.User
	err := a.b.DB().GetContext(ctx, &user, fmt.Sprintf("INSERT INTO pti_users (%s) VALUES ($1, $2) RETURNING %s;", userInsertFields, userFields), externalUserID, walletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return &user, nil
}

func (a *Activity) CheckPtiKYC(ctx context.Context, externalUserID string) error {
	return nil
}

func (a *Activity) CreatePtiWallet(ctx context.Context, args pti.CreateExternalWalletArgs) (*external.Wallet, error) {
	return a.external.CreateWallet(ctx, external.CreateWalletArgs{
		UserID:   args.UserID,
		WalletID: args.ID,
		Currency: args.Currency.String(),
	})
}

func (a *Activity) CreatePtiWalletLinkedAccount(ctx context.Context, args linkedaccounts.CreateArgs) (*linkedaccounts.LinkedAccount, error) {
	existing, err := a.b.LinkedAccounts().GetByProviderID(ctx, linkedaccounts.GetByProviderIDArgs{
		Provider:   args.Provider,
		ProviderID: args.ProviderID,
		WalletID:   args.WalletID,
	})
	if err != nil && !errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}

	return a.b.LinkedAccounts().Create(ctx, &args)
}
