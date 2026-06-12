package ops

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"gitlab.com/fynbos/backend/country"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/payments"
	"gitlab.com/fynbos/backend/providers/pti"
	"gitlab.com/fynbos/backend/providers/pti/external"
	"gitlab.com/fynbos/backend/slack"
	"gitlab.com/fynbos/backend/transactions"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"gitlab.com/fynbos/pacioli"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.uber.org/zap"
)

type Activity struct {
	b        Backends
	external external.Client
}

func NewActivity(b Backends, privateKey jwk.Key, baseURL, clientID string) *Activity {
	ptiExternalClient, err := external.NewWithOptions(
		external.WithBaseURL(baseURL),
		external.WithOTELLHTTPClient(),
		external.WithClientID(clientID),
		external.WithDerivedKeys(privateKey),
	)
	if err != nil {
		log.Fatalln(fmt.Errorf("%w Failed to create PTI external client: %s", pti.ErrInternal, err))
	}

	return &Activity{
		b:        b,
		external: ptiExternalClient,
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

	// now update the personal details needed for front-end
	externalUser, err := a.external.GetUser(ctx, usr.ExternalID)
	if err != nil {
		log.Error("pti activity (checkUserAssessmentAccepted): failed to fetch external user", zap.Error(err))
		return nil
	}

	dob, err := time.Parse(time.DateOnly, externalUser.DateOfBirth)
	if err != nil {
		log.Error("pti activity (checkUserAssessmentAccepted): failed to parse date of birth", zap.Error(err))
		return nil
	}

	args := kyc.IndividualDetails{
		WalletID:    walletID,
		FirstName:   externalUser.Name.First,
		LastName:    externalUser.Name.Last,
		DateOfBirth: dob,
		IPAddress:   "10.0.0.1",
	}
	if len(externalUser.Addresses) > 0 {
		args.Address = &kyc.Address{
			// State:   externalUser.Addresses[0].StateCode.Code,
			State:   externalUser.Addresses[0].StateCode,
			Line1:   externalUser.Addresses[0].Street,
			City:    externalUser.Addresses[0].City,
			ZipCode: externalUser.Addresses[0].PostalCode,
		}
	}

	_, err = a.b.KYC().UpdateIndividualDetails(ctx, args)
	if err != nil {
		log.Error("pti activity (checkUserAssessmentAccepted): failed to update individual details", zap.Error(err))
		return nil
	}

	err = a.b.KYC().SetKYCStatus(ctx, walletID, kyc.StatusLevel2)
	if err != nil {
		log.Error("pti activity (checkUserAssessmentAccepted): failed to set kyc status to level2", zap.Error(err))
		return nil
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
			DebitsMustNotExceedCredits: false,
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

func (a *Activity) CreatePTICard(ctx context.Context, walletID, tokenID string) (*linkedaccounts.LinkedAccount, error) {
	ptiUser, err := GetUser(ctx, a.b, walletID)
	if errors.Is(err, pti.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("pti create card: no external user found", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	externalCardRawJson, err := a.external.GetUsersPaymentInformation(ctx, ptiUser.ExternalID, tokenID)
	if err != nil {
		return nil, err
	}

	var d external.PaymentInformation
	err = json.Unmarshal(externalCardRawJson, &d)
	if err != nil {
		return nil, err
	}

	if d.Type != external.CardPaymentInformationType {
		return nil, temporal.NewNonRetryableApplicationError("pti create card: payment method is not a card", "ErrInternal", fmt.Errorf("%w payment method is not a card", pti.ErrInternal))
	}

	var cardInfo external.EncryptedCreditCardPaymentInformation
	err = json.Unmarshal(externalCardRawJson, &cardInfo)
	if err != nil {
		return nil, err
	}

	network := ""
	if cardInfo.CreditCardType != "" {
		network = cardInfo.CreditCardType
	}

	la, err := a.b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:            walletID,
		Name:                strings.TrimSpace(fmt.Sprintf("%s %s", network, cardInfo.CreditCardLast4)),
		Nickname:            strings.TrimSpace(fmt.Sprintf("%s %s", network, cardInfo.CreditCardLast4)),
		Provider:            pti.ProviderName,
		ProviderID:          tokenID,
		Mask:                cardInfo.CreditCardLast4,
		State:               linkedaccounts.Verified,
		Type:                pti.TypeCard,
		CanSend:             true,
		CanReceive:          true,
		SendCountry:         country.US,
		SendCurrency:        currency.USD,
		SendNetwork:         network,
		SendAvailability:    linkedaccounts.Immediate,
		ReceiveCountry:      country.US,
		ReceiveCurrency:     currency.USD,
		ReceiveNetwork:      network,
		ReceiveAvailability: linkedaccounts.Immediate,
	})
	if err != nil {
		return nil, err
	}

	return la, nil
}

func (a *Activity) CreatePtiBankAccount(ctx context.Context, args pti.CreateBankAccountArgs) (*external.BankAccountPaymentInformation, error) {
	ptiUser, err := GetUser(ctx, a.b, args.WalletID)
	if errors.Is(err, pti.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("save pti bank account: pti user not found", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	bank, err := a.external.CreateBankAccount(ctx, ptiUser.ExternalID, external.BankAccountPaymentInformation{
		Type:              "BANK_ACCOUNT",
		BankAccountNumner: args.AccountNumber,
		BankAccountType:   args.AccountType,
		BankRoutingNumber: args.RoutingNumber,
		// BankRoutingCheckDigit: args.RoutingNumberCheckDigit,
		AccountBankName: args.Bank,
	})
	if err != nil {
		return nil, temporal.NewNonRetryableApplicationError("create pti bank account fail, invalid bank account numbers", "ErrBankAccountInvalid", err)
	}

	return bank, nil
}

func (a *Activity) SavePtiBankAccount(ctx context.Context, walletID, externalPaymentMethodID string) (*linkedaccounts.LinkedAccount, error) {
	ptiUser, err := GetUser(ctx, a.b, walletID)
	if errors.Is(err, pti.ErrNotFound) {
		return nil, temporal.NewNonRetryableApplicationError("save pti bank account: pti user not found", "ErrNotFound", err)
	}
	if err != nil {
		return nil, err
	}

	data, err := a.external.GetUsersPaymentInformation(ctx, ptiUser.ExternalID, externalPaymentMethodID)
	if err != nil {
		return nil, err
	}

	var bank external.BankAccountPaymentInformation
	err = json.Unmarshal(data, &bank)
	if err != nil {
		return nil, err
	}

	mask := bank.BankAccountNumner
	if len(mask) > 4 {
		mask = bank.BankAccountNumner[len(mask)-4:]
	}

	la, err := a.b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:            walletID,
		Name:                strings.TrimSpace(fmt.Sprintf("%s %s", bank.AccountBankName, mask)),
		Nickname:            strings.TrimSpace(fmt.Sprintf("%s %s", bank.AccountBankName, mask)),
		Provider:            pti.ProviderName,
		ProviderID:          externalPaymentMethodID,
		Mask:                mask,
		State:               linkedaccounts.Verified,
		Type:                pti.TypeBank,
		CanSend:             true,
		CanReceive:          true,
		SendCountry:         country.US,
		SendCurrency:        currency.USD,
		SendNetwork:         "ACH",
		SendAvailability:    linkedaccounts.Few,
		ReceiveCountry:      country.US,
		ReceiveCurrency:     currency.USD,
		ReceiveNetwork:      "ACH",
		ReceiveAvailability: linkedaccounts.Few,
	})
	if err != nil {
		return nil, err
	}

	return la, nil
}

func getWalletID(ctx context.Context, b Backends, externalUserID string) (string, error) {
	var walletID string
	err := b.DB().GetContext(ctx, &walletID, "SELECT wallet_id FROM pti_users WHERE external_id=$1;", externalUserID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w %s", pti.ErrNotFound, err)
	} else if err != nil {
		return "", fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	return walletID, nil
}

func (a *Activity) GetWalletFromPTIUser(ctx context.Context, externalUserID string) (string, error) {

	walletID, err := getWalletID(ctx, a.b, externalUserID)
	if errors.Is(err, pti.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrNotFound", fmt.Errorf("%w No wallet found for pti user", pti.ErrNotFound))
	}
	if err != nil {
		return "", err
	}

	return walletID, nil
}

func (a *Activity) CreateTransaction(ctx context.Context, la *linkedaccounts.LinkedAccount, amount currency.Amount, externalID, note string) error {
	if amount.Currency != currency.USD {
		return temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", pti.ErrInternal))
	}

	wallet, err := a.b.Wallets().Get(ctx, la.WalletID)
	if errors.Is(err, pti.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Wallet not found", "ErrNotFound", fmt.Errorf("%w No wallet found for pti user", pti.ErrNotFound))
	}
	if err != nil {
		return err
	}

	las, err := a.b.LinkedAccounts().ListBalances(ctx, la.WalletID)
	if err != nil {
		return err
	}

	var balance *linkedaccounts.LinkedAccount
	for _, l := range las {
		if l.Provider == pti.ProviderName && l.Type == pti.AccTypeBalance {
			balance = &l
			break
		}
	}
	if balance == nil {
		return temporal.NewNonRetryableApplicationError("PTI USD balance account not found", "ErrInternal", fmt.Errorf("%w balance account not found", pti.ErrInternal))
	}

	_, err = a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		WalletID:                wallet.ID,
		ForeignID:               externalID,
		ForeignType:             transactions.TransactionTypeDeposit,
		Provider:                pti.ProviderName,
		State:                   transactions.StatePending,
		Source:                  wallet.AddressString(),
		Destination:             wallet.AddressString(),
		Title:                   "Deposit",
		DestinationIdentity:     wallet.ID,
		DestinationIdentityType: payments.IdentityTypeWalletID.String(),
		Amount:                  amount,
		ProviderFee:             nil,
		LinkedAccountTitle:      "US Balance",
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: balance.ID,
				ForeignID:       externalID,
				Amount:          amount,
				Type:            transactions.TransferTypeCreditBalance,
				State:           transactions.StatePending,
			},
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) UpdateTransaction(ctx context.Context, transaction *transactions.Transaction, externalTxID string) error {
	transfers, err := a.b.Transactions().ListTransfers(ctx, transaction.ID)
	if err != nil || len(transfers) != 1 { // deposit only
		return errors.New("not a deposit transaction")
	}
	transferID := transfers[0].ID

	err = a.b.Transactions().SetTransferForeignID(ctx, transferID, externalTxID)
	if err != nil {
		return err
	}

	return a.b.Transactions().SetTransactionForeignID(ctx, transaction.ID, externalTxID)
}

func (a *Activity) GetTransactionByForeignID(ctx context.Context, walletID string, foreignID string) (*transactions.Transaction, error) {
	return a.b.Transactions().GetTransactionByForeignID(ctx, walletID, foreignID)
}

func (a *Activity) CheckPaymentState(ctx context.Context, externalTxID string) error {
	p, err := a.b.Payments().Lookup(ctx, externalTxID)
	if err != nil {
		return temporal.NewNonRetryableApplicationError("PTI USD Payment Not found", "ErrInternal", fmt.Errorf("%w Fiant payment not found %s", pti.ErrNotFound, externalTxID))
	}
	if p.State == payments.StateCompleted {
		return temporal.NewNonRetryableApplicationError("PTI USD payment completed ", "ErrInternal", fmt.Errorf("%w Fiant payment previously completed %s", pti.ErrAssessmentFailed, externalTxID))
	}
	return nil
}

func (a *Activity) SettleTransaction(ctx context.Context, transactionID, walletID string, amount currency.Amount) (string, error) {
	existingTransaction, err := a.b.Transactions().GetTransactionByForeignID(ctx, walletID, transactionID)
	if err != nil && !errors.Is(err, transactions.ErrNotFound) {
		return "", err
	}
	if existingTransaction != nil {
		if existingTransaction.State == transactions.StateCompleted {
			return "", temporal.NewNonRetryableApplicationError("PTI USD payment deposit", "ErrInternal", fmt.Errorf("%w Fiant payment previously completed %s", pti.ErrAssessmentFailed, transactionID))
		}
		if err := a.b.Transactions().SetTransactionState(ctx, existingTransaction.ID, transactions.StateCompleted); err != nil {
			return "", err
		}

		transfers, err := a.b.Transactions().ListTransfers(ctx, existingTransaction.ID)
		if err != nil || len(transfers) != 1 { // deposit only
			return "", errors.New("not a deposit transaction")
		}
		transferID := transfers[0].ID

		err = a.b.Transactions().SetTransferState(ctx, transferID, transactions.StateCompleted)
		if err != nil {
			return "", err
		}

		return existingTransaction.ID, nil
	}

	if amount.Currency != currency.USD {
		return "", temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", pti.ErrInternal))
	}

	wallet, err := a.b.Wallets().Get(ctx, walletID)
	if errors.Is(err, pti.ErrNotFound) {
		return "", temporal.NewNonRetryableApplicationError("Wallet not found", "ErrNotFound", fmt.Errorf("%w No wallet found for pti user", pti.ErrNotFound))
	}
	if err != nil {
		return "", err
	}

	las, err := a.b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return "", err
	}

	var eurBalance *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance {
			eurBalance = &la
			break
		}
	}
	if eurBalance == nil {
		return "", temporal.NewNonRetryableApplicationError("PTI USD balance account not found", "ErrInternal", fmt.Errorf("%w Fiant USD balance account not found", pti.ErrInternal))
	}

	tx, err := a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		ID:                      transactionID,
		WalletID:                walletID,
		ForeignID:               transactionID,
		ForeignType:             transactions.TransactionTypeDeposit,
		Provider:                pti.ProviderName,
		State:                   transactions.StateCompleted,
		Source:                  wallet.AddressString(),
		Destination:             wallet.AddressString(),
		Title:                   "Deposit",
		DestinationIdentity:     walletID,
		DestinationIdentityType: payments.IdentityTypeWalletID.String(),
		Amount:                  amount,
		ProviderFee:             nil,
		LinkedAccountTitle:      "US Balance",
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: eurBalance.ID,
				ForeignID:       transactionID,
				Amount:          amount,
				Type:            transactions.TransferTypeCreditBalance,
				State:           transactions.StateCompleted,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return tx, nil
}

func (a *Activity) FinalizePTIDeposit(ctx context.Context, id, walletID string, amount currency.Amount) error {
	if amount.Currency != currency.USD {
		return temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", pti.ErrInternal))
	}

	las, err := a.b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return err
	}

	var USDBalance *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance {
			USDBalance = &la
			break
		}
	}
	if USDBalance == nil {
		return temporal.NewNonRetryableApplicationError("PTI USD balance account not found", "ErrInternal", fmt.Errorf("%w PTI USD balance account not found", pti.ErrInternal))
	}

	opsAcc := pti.USDOpsAccount
	ledger := pti.LedgerIDUSD
	tx, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              id,
			Amount:          amount.Value,
			CreditAccountID: USDBalance.ID,
			DebitAccountID:  opsAcc,
			Pending:         false,
			Code:            1,
			Ledger:          ledger,
		},
	})
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	if len(tx) > 0 {
		if tx[0].Code == pacioli.TransferExists {
			return nil
		}
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return fmt.Errorf("%w insufficient balance code (%s)", pti.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return fmt.Errorf("%w non success code (%s)", pti.ErrInternal, tx[0].Code.String())
		}
	}

	return nil
}

func (a *Activity) PTIDeposit(ctx context.Context, ptiArgs pti.TransactionArgs) (string, error) {
	return DepositToWallet(ctx, a.b, a.external, ptiArgs)

}

func (a *Activity) ReservePTIBalance(ctx context.Context, walletID, id string) error {
	tx, err := a.b.Transactions().GetTransaction(ctx, walletID, id)
	if errors.Is(err, transactions.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Transaction not found", "ErrInternal", fmt.Errorf("%w transaction not found", pti.ErrInternal))
	}
	if tx.Amount.Currency != currency.USD {
		return temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", pti.ErrInternal))
	}

	transfers, err := a.b.Transactions().ListTransfers(ctx, tx.ID)
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}

	if len(transfers) < 1 {
		return fmt.Errorf("%w unable to reserve balance. No linked account specified", pti.ErrInternal)
	}

	timeout := time.Hour * 24 * 365 // Pending transfers must have a timeout.
	var fee int64 = 0
	if tx.ProviderFee != nil {
		fee = tx.ProviderFee.Value
	}

	amountWithFee := currency.Amount{
		Value:    tx.Amount.Value + fee,
		Currency: tx.Amount.Currency,
		Scale:    tx.Amount.Scale,
	}

	_, err = ReserveBalance(ctx, a.b, transfers[0].LinkedAccountID, tx.ID, amountWithFee, timeout)
	return err
}

func (a *Activity) FinalizePTIBalance(ctx context.Context, id, walletID string) error {
	tx, err := a.b.Transactions().GetTransaction(ctx, walletID, id)
	if errors.Is(err, transactions.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Transaction not found", "ErrInternal", fmt.Errorf("%w transaction not found", pti.ErrInternal))
	}
	if tx.Amount.Currency != currency.USD {
		return temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", pti.ErrInternal))
	}

	return FinaliseReserve(ctx, a.b, tx.ID)
}

func (a *Activity) UpdatePaymentState(ctx context.Context, paymentID string, state payments.State) error {
	_, err := a.b.DB().ExecContext(ctx, "UPDATE payments SET state=$1, updated_at=now() where id=$2", payments.StateCompleted, paymentID)
	if err != nil {
		return err
	}
	return nil
}

func (a *Activity) UpdateTransactionState(ctx context.Context, walletID, transactionID string, state transactions.State) error {
	info := activity.GetInfo(ctx)
	if info.Attempt == 1 && state == transactions.StateFailed {
		slack.SendToChannel(ctx, slack.ChannelError, "wallet-info-bot", fmt.Sprintf("Pti withdrawal failed. %s/wallet/%s/transactions/%s", env.AdminURL(), walletID, transactionID))
	}

	transfers, err := a.b.Transactions().ListTransfers(ctx, transactionID)
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	for _, t := range transfers {
		err = a.b.Transactions().SetTransferState(ctx, t.ID, state)
		if err != nil {
			return fmt.Errorf("%w %s", pti.ErrInternal, err)
		}
	}

	return a.b.Transactions().SetTransactionState(ctx, transactionID, state)
}

func (a *Activity) RollbackPTIBalance(ctx context.Context, id, walletID string) error {
	tx, err := a.b.Transactions().GetTransaction(ctx, walletID, id)
	if errors.Is(err, transactions.ErrNotFound) {
		return temporal.NewNonRetryableApplicationError("Transaction not found", "ErrInternal", fmt.Errorf("%w transaction not found", pti.ErrInternal))
	}
	if tx.Amount.Currency != currency.USD {
		return temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", pti.ErrInternal))
	}

	return RollbackReserve(ctx, a.b, tx.ID)
}

func (a *Activity) ReturnTransaction(ctx context.Context, originalTransactionID, walletID string, amount currency.Amount) (string, error) {
	originalTransaction, err := a.b.Transactions().GetTransactionByForeignID(ctx, walletID, originalTransactionID)
	if err != nil && !errors.Is(err, transactions.ErrNotFound) {
		return "", err
	}

	las, err := a.b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return "", err
	}
	var balance *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance {
			balance = &la
			break
		}
	}
	if balance == nil {
		return "", temporal.NewNonRetryableApplicationError("PTI USD balance account not found", "ErrInternal", fmt.Errorf("%w Fiant USD balance account not found", pti.ErrInternal))
	}

	if originalTransaction.Type == transactions.TransactionTypeDeposit && originalTransaction.State == transactions.StatePending {
		// if the original transaction is still pending (deposit only) we can just fail it without creating a return transaction
		err = a.b.Transactions().SetTransactionState(ctx, originalTransaction.ID, transactions.StateFailed)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	returnedTransaction, err := a.b.Transactions().GetTransactionByForeignID(ctx, walletID, originalTransaction.ID)
	if err != nil && !errors.Is(err, transactions.ErrNotFound) {
		return "", err
	}

	if returnedTransaction != nil {
		return "already has a return transaction", temporal.NewNonRetryableApplicationError("already has a return transaction", "ErrInternal", fmt.Errorf("%w already has a return transaction", pti.ErrInternal))
	}

	var title, returnSource, returnDestination string
	var transferType transactions.TransferType
	switch originalTransaction.Type {
	case transactions.TransactionTypeDeposit:
		title = "ACH return for deposit"
		returnSource = originalTransaction.Destination
		returnDestination = originalTransaction.Source
		transferType = transactions.TransferTypeDebitBalance

	case transactions.TransactionTypeWithdrawal:
		title = "ACH return for withdrawal"
		returnSource = originalTransaction.Source
		returnDestination = originalTransaction.Destination
		transferType = transactions.TransferTypeCreditBalance
	default:
		return "", temporal.NewNonRetryableApplicationError("Unsupported transaction type", "ErrInternal", fmt.Errorf("%w unsupported transaction type for return: %s", pti.ErrInternal, originalTransaction.Type))
	}

	returnTransactionID, err := a.b.Transactions().CreateTransaction(ctx, transactions.CreateTransactionArgs{
		WalletID:                walletID,
		ForeignID:               originalTransaction.ID,
		ForeignType:             transactions.TransactionTypeReturn,
		Provider:                pti.ProviderName,
		State:                   transactions.StateCompleted,
		Source:                  returnSource,
		Destination:             returnDestination,
		Title:                   title,
		DestinationIdentity:     walletID,
		DestinationIdentityType: payments.IdentityTypeWalletID.String(),
		Amount:                  amount,
		ProviderFee:             nil,
		LinkedAccountTitle:      "US Balance",
		Transfers: []transactions.TransferArgs{
			{
				LinkedAccountID: balance.ID,
				ForeignID:       originalTransactionID,
				Amount:          amount,
				Type:            transferType,
				State:           transactions.StateCompleted,
			},
		},
	})
	if err != nil {
		return "", err
	}

	return returnTransactionID, nil
}

func (a *Activity) PostTransfer(ctx context.Context, transactionID, walletID string) error {
	returnTransaction, err := a.b.Transactions().GetTransaction(ctx, walletID, transactionID)
	if err != nil {
		return err
	}

	originalTransaction, err := a.b.Transactions().GetTransaction(ctx, walletID, returnTransaction.ForeignID)
	if err != nil {
		return err
	}

	if returnTransaction.Amount.Currency != currency.USD {
		return temporal.NewNonRetryableApplicationError("Invalid currency", "ErrInternal", fmt.Errorf("%w invalid currency", pti.ErrInternal))
	}

	las, err := a.b.LinkedAccounts().ListBalances(ctx, walletID)
	if err != nil {
		return err
	}

	var USDBalance *linkedaccounts.LinkedAccount
	for _, la := range las {
		if la.Provider == pti.ProviderName && la.Type == pti.AccTypeBalance {
			USDBalance = &la
			break
		}
	}
	if USDBalance == nil {
		return temporal.NewNonRetryableApplicationError("PTI USD balance account not found", "ErrInternal", fmt.Errorf("%w PTI USD balance account not found", pti.ErrInternal))
	}

	opsAcc := pti.USDOpsAccount
	ledger := pti.LedgerIDUSD

	var creditAccountID, debitAccountID string
	switch originalTransaction.Type {
	case transactions.TransactionTypeDeposit:
		creditAccountID = opsAcc
		debitAccountID = USDBalance.ID
	case transactions.TransactionTypeWithdrawal:
		creditAccountID = USDBalance.ID
		debitAccountID = opsAcc
	default:
		return temporal.NewNonRetryableApplicationError("Unsupported transaction type", "ErrInternal", fmt.Errorf("%w unsupported transaction type for return: %s", pti.ErrInternal, originalTransaction.Type))
	}

	tx, err := a.b.Pacioli().CreateTransfers(ctx, []pacioli.CreateTransferArgs{
		{
			ID:              returnTransaction.ID,
			Amount:          returnTransaction.Amount.Value,
			CreditAccountID: creditAccountID,
			DebitAccountID:  debitAccountID,
			Pending:         false,
			Code:            1,
			Ledger:          ledger,
		},
	})
	if err != nil {
		return fmt.Errorf("%w %s", pti.ErrInternal, err)
	}
	if len(tx) > 0 {
		if tx[0].Code == pacioli.TransferExists {
			return nil
		}
		if tx[0].Code == pacioli.TransferExceedsCredits || tx[0].Code == pacioli.TransferExceedsDebits || tx[0].Code == pacioli.TransferExceedsPendingTransferAmount {
			return fmt.Errorf("%w insufficient balance code (%s)", pti.ErrInsufficientBalance, tx[0].Code.String())
		}
		if tx[0].Code != 0 {
			return fmt.Errorf("%w non success code (%s)", pti.ErrInternal, tx[0].Code.String())
		}
	}

	return nil
}
