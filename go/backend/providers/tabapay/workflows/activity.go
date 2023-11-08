package workflows

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/basistheory"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/tabapay"
	"gitlab.com/fynbos/backend/providers/tabapay/external"
	external_client "gitlab.com/fynbos/backend/providers/tabapay/external/client"
	mock_external_client "gitlab.com/fynbos/backend/providers/tabapay/external/client/mock"
	"gitlab.com/fynbos/backend/wallets"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
		SubClientID:            os.Getenv("TABAPAY_SUB_CLIENT_ID"),
		Transport: otelhttp.NewTransport(
			httplogger.NewTransport(http.DefaultTransport, cb, nil),
		),
	}

	var externalClient external.Client
	if env.IsLocal() {
		externalClient = mock_external_client.SetupDevMock(nil)
	} else {
		c, err := external_client.New(clientArgs)
		if err != nil {
			log.Fatal("Failed to create Tabapay activity.", zap.Error(err))
		}
		externalClient = c
	}

	return &Activity{b: &backends{
		cb, externalClient,
	}}
}

func (a *Activity) QueryCard(ctx context.Context, args QueryCard) (*external.QueryCardResponse, error) {
	kyc, err := a.b.KYC().GetIndividualDetails(ctx, args.WalletID)
	if err != nil {
		return nil, err
	}

	queryArgs := external.QueryCardArgs{
		Card: &external.Card{
			AccountNumber:  args.CardNumber,
			SecurityCode:   args.CVV,
			ExpirationDate: args.ExpirationDate,
		},
		Owner: &external.Owner{
			Name: external.Name{
				First: kyc.FirstName,
				Last:  kyc.LastName,
			},
		},
	}
	if args.AVS {
		ctry, err := country.Country(kyc.Address.CountryCode).Numeric()
		if err != nil {
			err = fmt.Errorf("%w invalid country=%s", tabapay.ErrInternal, kyc.Address.CountryCode)
			return nil, temporal.NewNonRetryableApplicationError("tabapay: Unsupported country.", "ErrUnsupportedCountry", err)
		}
		if !country.Country(kyc.Address.CountryCode).IsSupported() {
			err = fmt.Errorf("%w unsupported country=%s", tabapay.ErrInternal, kyc.Address.CountryCode)
			return nil, temporal.NewNonRetryableApplicationError("tabapay: Unsupported country.", "ErrUnsupportedCountry", err)
		}
		queryArgs.AVSCheck = true
		queryArgs.Owner.Address = &external.Address{
			Line1:   kyc.Address.Line1,
			ZipCode: kyc.Address.ZipCode,
			Country: ctry,
		}
	}

	resp, err := a.b.External().QueryCard(ctx, queryArgs)
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}
	if resp.SC == http.StatusMultiStatus {
		err = fmt.Errorf("%w MultiStatus response. Tabapay error code=%s", tabapay.ErrMultiStatus, resp.EC)
		return nil, temporal.NewApplicationErrorWithCause("tabapay: unavailable", "ErrMultiStatus", err)
	}

	return resp, nil
}

func (a *Activity) CreateExternalCard(ctx context.Context, args CreateExternalCardArgs) (*external.CreateAccountResponse, error) {
	owner, err := a.b.KYC().GetIndividualDetails(ctx, args.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	externalArgs := external.CreateAccountArgs{
		RejectDuplicateCard:  args.RejectDuplicateCard,
		OKToAddDuplicateCard: !args.RejectDuplicateCard,
		ReferenceID:          args.ReferenceID,
		Card: external.Card{
			AccountNumber:  args.CardNumber,
			ExpirationDate: args.ExpirationDate,
		},
		Owner: external.Owner{
			Name: external.Name{
				First: owner.FirstName,
				Last:  owner.LastName,
			},
		},
	}

	if args.AddAddress && owner.Address == nil {
		log.Warn("no address found in kyc. Address will not be submitted to Tabapay.", zap.String("walletID", args.WalletID))
	} else if args.AddAddress {
		ctry, err := country.Country(owner.Address.CountryCode).Numeric()
		if err != nil {
			err = fmt.Errorf("%w invalid country=%s", tabapay.ErrInternal, owner.Address.CountryCode)
			return nil, temporal.NewNonRetryableApplicationError("tabapay: Unsupported country.", "ErrUnsupportedCountry", err)
		}
		if !country.Country(owner.Address.CountryCode).IsSupported() {
			err = fmt.Errorf("%w unsupported country=%s", tabapay.ErrInternal, owner.Address.CountryCode)
			return nil, temporal.NewNonRetryableApplicationError("tabapay: Unsupported country.", "ErrUnsupportedCountry", err)
		}
		stateParts := strings.Split(owner.Address.State, "-")
		state := stateParts[0]
		if len(stateParts) == 2 {
			state = stateParts[1]
		}

		externalArgs.Owner.Address = &external.Address{
			Line1:   owner.Address.Line1,
			Line2:   owner.Address.Line2,
			City:    owner.Address.City,
			State:   state,
			ZipCode: owner.Address.ZipCode,
			Country: ctry,
		}
	}

	resp, err := a.b.External().CreateAccount(ctx, externalArgs)
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	return resp, nil
}

func (a *Activity) CreateLinkedCard(ctx context.Context, args linkedaccounts.CreateArgs) (*linkedaccounts.LinkedAccount, error) {
	la, err := a.b.LinkedAccounts().Create(ctx, &args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", tabapay.ErrInternal, err)
	}

	if la.State != linkedaccounts.Verified {
		_, err = a.b.LinkedAccounts().CreateReviews(ctx, []linkedaccounts.CreateReviewArgs{
			{
				LinkedAccountID: la.ID,
				State:           la.State,
			},
		})
	}
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

func (a *Activity) CreateBasisTheoryCard(ctx context.Context, args basistheory.CreateCardArgs) (*basistheory.Card, error) {
	card, err := a.b.BasisTheory().CreateCard(ctx, args)
	if err != nil {
		return nil, err
	}

	return card, nil
}

func (a *Activity) ListLinkedAccountsByProviderID(ctx context.Context, provider, providerID string) ([]linkedaccounts.LinkedAccount, error) {
	las, err := a.b.LinkedAccounts().ListByProviderID(ctx, provider, providerID)
	if err != nil {
		return nil, err
	}

	return las, nil
}

func (a *Activity) GetWallet(ctx context.Context, id string) (*wallets.Wallet, error) {
	return a.b.Wallets().Get(ctx, id)
}
