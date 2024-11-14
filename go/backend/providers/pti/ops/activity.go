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
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	httplogger "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/external"
	external_mock "gitlab.com/fynbos/backend/providers/pti/external/mock"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/pacioli"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.temporal.io/sdk/temporal"
)

type Activity struct {
	b        Backends
	external external.Client
}

func NewActivity(b Backends, privateKey jwk.Key) *Activity {
	var ex external.Client
	if env.IsTest() {
		ex = external_mock.SetupDevMock(nil)
	} else if env.IsLocal() {
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
	u, err := GetUser(ctx, a.b, walletID)
	if err != nil {
		if errors.Is(err, pti.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return u, err
}

func (a *Activity) CreatePtiUser(ctx context.Context, walletID string) (string, error) {
	var emails []external.Email
	var phones []external.Phone
	usrs, err := a.b.Users().ListUsers(ctx, walletID)
	if err != nil {
		return "", err
	}

	var firstName, lastName string

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

		if firstName == "" {
			firstName = usr.FirstName
		}

		if lastName == "" {
			lastName = usr.LastName
		}
	}

	return a.external.CreateUser(ctx, external.CreateUserArgs{
		ID:   uuid.NewString(),
		Type: "PERSON",
		Name: external.Name{
			First: firstName,
			Last:  lastName,
		},
		Emails: emails,
		Phones: phones,
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
	if existing != nil && existing.DeletedAt.Time.IsZero() {
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

	receiverPTIUser, err := GetUser(ctx, a.b, p.Receiver.WalletID)
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
		SessionID:  p.ID,
		Amount:     p.SenderAmount.Float64(),
		USDValue:   p.SenderAmount.Float64(),
		TransactionTotal: external.Total{
			Fee: external.Cost{
				Amount:   0,
				Currency: p.SenderAmount.Currency.String(),
			},
			Subtotal: external.Cost{
				Amount:   p.SenderAmount.Float64(),
				Currency: p.SenderAmount.Currency.String(),
			},
			Total: external.Cost{
				Amount:   p.SenderAmount.Float64(),
				Currency: p.SenderAmount.Currency.String(),
			},
		},
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
		Date:           time.Now().Format(time.RFC3339),
	})
	if errors.Is(err, external.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("PTI user not found", "ErrNotFound", err)
	}
	if errors.Is(err, external.ErrUnprocessableEntity) {
		return nil, temporal.NewApplicationError("PTI unable to process transfer", "ErrUnprocessableEntity", err)
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

func (a *Activity) CreatePTIBalanceAccount(ctx context.Context, id string) error {
	accs, err := a.b.Pacioli().ConfigureAccounts(ctx, []pacioli.ConfigureAccountArgs{
		{
			ID:                         id,
			LedgerID:                   pti.LedgerIDUSD,
			Code:                       1,
			DebitsMustNotExceedCredits: true,
			CreditsMustNotExceedDebits: false,
		},
	})
	if err != nil {
		return err
	}

	if len(accs) == 0 {
		// No error codes to speak of
		return nil
	}

	if accs[0].Code != pacioli.AccountOK && accs[0].Code != pacioli.AccountExists {
		return fmt.Errorf("%w failed to setup account status(%s)", pti.ErrInternal, accs[0].Code)
	}

	return nil
}
