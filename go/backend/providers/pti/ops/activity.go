package ops

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/external"
	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/temporal"
)

type Activity struct {
	b        Backends
	external external.Client
}

func NewActivity(b Backends, privateKey jwk.Key) *Activity {
	var ex external.Client
	if env.IsLocal() {
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			log.Fatalln(err)
		}

		ptiPrivateKey, err := jwk.FromRaw(privateKey)
		if err != nil {
			log.Fatalln(err)
		}
		ex = external.New(external.ClientArgs{
			Transport: &http.Client{
				Transport: otelhttp.NewTransport(
					httplogger.NewTransport(http.DefaultTransport, b, nil),
				),
			},
			ClientID:   "LOCAL",
			PrivateKey: ptiPrivateKey,
		})
	} else {
		ex = external.New(external.ClientArgs{
			Transport: &http.Client{
				Transport: otelhttp.NewTransport(
					httplogger.NewTransport(http.DefaultTransport, b, nil),
				),
			},
			ClientID:   os.Getenv("PTI_CLIENT_ID"),
			PrivateKey: privateKey,
		})
	}

	return &Activity{
		b:        b,
		external: ex,
	}
}

func (a *Activity) GetPtiUser(ctx context.Context, walletID string) (*pti.User, error) {
	return GetUser(ctx, a.b, walletID)
}

func (a *Activity) CreatePtiUser(ctx context.Context, walletID string) (string, error) {
	kycData, err := a.b.KYC().GetIndividualDetails(ctx, walletID)
	if errors.Is(err, kyc.ErrNoKYCInfo) {
		return "", temporal.NewNonRetryableApplicationError("No KYC data for wallet.", "ErrNotFound", err)
	}

	var addresses []external.Address
	if kycData.Address != nil {
		addresses = append(addresses, external.Address{
			Street: kycData.Address.Line1,
			City:   kycData.Address.City,
			StateCode: external.StateCode{
				Code: kycData.Address.State,
			},
			Country: external.CountryCode{
				Code: kycData.Address.CountryCode,
			},
			PostalCode: kycData.Address.ZipCode,
			Default:    true,
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
		Type:        "PERSON",
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

func (a *Activity) StartUserAssessment(ctx context.Context, walletID string, scenarioID string) (string, error) {
	usr, err := GetUser(ctx, a.b, walletID)
	if err != nil {
		return "", err
	}

	return a.external.StartUserAssessment(ctx, external.StartUserAssessmentArgs{
		ID:         usr.ExternalID,
		ScenarioID: scenarioID,
		Type:       "PERSON",
	})
}

func (a *Activity) CheckUserAssessmentAccepted(ctx context.Context, walletID string) error {
	usr, err := GetUser(ctx, a.b, walletID)
	if err != nil {
		return err
	}

	assessment, err := a.external.GetUserAssessment(ctx, usr.ExternalID)
	if err != nil {
		return err
	}

	if assessment.Assessment != "ACCEPTED" {
		return fmt.Errorf("%w", pti.ErrAssessmentFailed)
	}

	return nil
}

func (a *Activity) CreatePtiWallet(ctx context.Context, args pti.CreateExternalWalletArgs) (*external.Wallet, error) {
	wallets, err := a.external.ListWallets(ctx, args.UserID)
	if err != nil {
		return nil, err
	}

	for _, w := range wallets {
		if w.WalletID == args.ID && w.Currency == args.Currency.String() {
			return &w, nil
		}
	}

	return a.external.CreateWallet(ctx, external.CreateWalletArgs{
		UserID:   args.UserID,
		WalletID: args.ID,
		Currency: args.Currency.String(),
		Type:     "WALLET",
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

func (a *Activity) PTIValidatePayment(ctx context.Context, paymentID string) error {
	p, err := a.b.Payments().Lookup(ctx, paymentID)
	if errors.Is(err, payments.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Payment not found", "ErrNotFound", err)
	}
	if err != nil {
		return err
	}

	if p.State == payments.StateProcessing {
		return temporal.NewNonRetryableApplicationError("Payment not in processing state", "ErrInternal", err)
	}

	return nil
}

func (a *Activity) CreateWalletTransfer(ctx context.Context, paymentID, requestID string) (*external.IDResponse, error) {
	p, err := a.b.Payments().Lookup(ctx, paymentID)
	if errors.Is(err, payments.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("Payment not found", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	senderPTIUser, err := GetUser(ctx, a.b, p.Sender.WalletID)
	if errors.Is(err, pti.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("PTI user not found for sender", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	senderPTIWalletLinkedAccount, err := a.b.LinkedAccounts().Get(ctx, p.SenderAccount)
	if errors.Is(err, pti.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("PTI linked account not found for sender", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	receiverPTIUser, err := GetUser(ctx, a.b, p.Sender.WalletID)
	if errors.Is(err, pti.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("PTI user not found for sender", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	receiverPTIWalletLinkedAccount, err := a.b.LinkedAccounts().Get(ctx, p.ReceiverAccount)
	if errors.Is(err, pti.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("PTI linked account not found for receiver", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	trxResp, err := a.external.CreateTransfer(ctx, external.TransferArgs{
		RequestID:  requestID,
		ScenarioID: pti.ScenarioTransfer,
		Amount:     p.SenderAmount.Float64(),
		USDValue:   p.SenderAmount.Float64(),
		Initiator: external.User{
			ID:   senderPTIUser.ExternalID,
			Type: "PERSON",
		},
		SourceTransferMethod: external.WalletPaymentMethod{
			PaymentMethodType: "WALLET",
			PaymentInformation: external.WalletType{
				ID:   senderPTIWalletLinkedAccount.ProviderID,
				Type: "WALLET",
			},
		},
		Destination: external.User{
			ID:   receiverPTIUser.ExternalID,
			Type: "PERSON",
		},
		DestinationTransferMethod: external.WalletPaymentMethod{
			PaymentMethodType: "WALLET",
			PaymentInformation: external.WalletType{
				ID:   receiverPTIWalletLinkedAccount.ProviderID,
				Type: "WALLET",
			},
		},
		Type:           "TRANSFER",
		DisableWebhook: true, // our workflows keep the context
	})
	if errors.Is(err, external.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("PTI user not found", "ErrNotFound", err)
	}
	if errors.Is(err, external.ErrUnprocessableEntity) {
		return nil, temporal.NewNonRetryableApplicationError("PTI unable to process assessment", "ErrInternal", err)
	}

	return trxResp, err
}

func (a *Activity) SavePTITransaction(ctx context.Context, paymentID, requestID string) error {
	res, err := a.b.DB().ExecContext(ctx, "INSERT INTO pti_transactions (payment_id, external_request_id) VALUES ($1,$2);", paymentID, requestID)
	if err != nil {
		return err
	}

	if rows, _ := res.RowsAffected(); rows < 1 {
		return fmt.Errorf("%w Failed to insert pti_transaction", pti.ErrInternal)
	}

	return nil
}

func (a *Activity) GetPTITransactionByPaymentID(ctx context.Context, paymentID string) (*external.TransactionStatus, error) {
	var externalReqeustID string
	err := a.b.DB().GetContext(ctx, &externalReqeustID, "SELECT external_request_id FROM pti_transactions WHERE payment_id=$1 ORDER BY created_at DESC LIMIT 1;", paymentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, temporal.NewNonRetryableApplicationError("PTI transaction not found for payment", "ErrNotFound", err)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return a.external.GetTransaction(ctx, externalReqeustID)
}

func (a *Activity) GetPTITransaction(ctx context.Context, requestID string) (*external.TransactionStatus, error) {
	return a.external.GetTransaction(ctx, requestID)
}
