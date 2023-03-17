package ops

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/db"
	"gitlab.com/fynbos/backend/linkedaccounts"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/providers/machnet/external"
	"gitlab.com/fynbos/backend/statements"
	"gitlab.com/fynbos/backend/transactions"
)

const (
	UserEventsChannel                = "machnet_user_events"
	TransactionEventsChannel         = "machnet_transaction_events"
	TransactionDeliveryEventsChannel = "machnet_delivery_events"
)

func CreateUser(ctx context.Context, b Backends, args machnet.CreateArgs) (*machnet.User, error) {
	var user machnet.User
	err := b.DB().GetContext(
		ctx,
		&user,
		"INSERT INTO machnet_users (id, wallet_id, kyc_status) VALUES ($1, $2, $3) RETURNING id, wallet_id, kyc_status, created_at, updated_at;",
		args.ExternalID,
		args.WalletID,
		machnet.KYCStatusInProgress,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &user, nil
}

func GetKYCStatus(ctx context.Context, b Backends, walletID string) (*machnet.UserKYC, error) {
	u, err := GetUserByWalletID(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	if u.KYCStatus != machnet.KYCStatusRetry {
		return &machnet.UserKYC{
			User: *u,
		}, nil
	}

	vs, err := b.External().GetVerificationStatus(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	var failed []string

	if vs.CipInfo.PhoneNumber == external.StatusFailed {
		failed = append(failed, "phoneNumber")
	}
	if vs.CipInfo.Email == external.StatusFailed {
		failed = append(failed, "email")
	}
	if vs.CipInfo.DateOfBirth == external.StatusFailed {
		failed = append(failed, "dateOfBirth")
	}
	if vs.CipInfo.Gender == external.StatusFailed {
		failed = append(failed, "gender")
	}
	if vs.CipInfo.FirstName == external.StatusFailed {
		failed = append(failed, "firstName")
	}
	if vs.CipInfo.LastName == external.StatusFailed {
		failed = append(failed, "lastName")
	}
	if vs.CipInfo.State == external.StatusFailed ||
		vs.CipInfo.ZipCode == external.StatusFailed ||
		vs.CipInfo.City == external.StatusFailed ||
		vs.CipInfo.Country == external.StatusFailed ||
		vs.CipInfo.AddressLine1 == external.StatusFailed ||
		vs.CipInfo.AddressLine2 == external.StatusFailed {
		failed = append(failed, "address")
	}

	return &machnet.UserKYC{
		User:         *u,
		FailedFields: failed,
	}, nil
}

func GetUserByWalletID(ctx context.Context, b Backends, walletID string) (*machnet.User, error) {
	var user machnet.User
	err := b.DB().GetContext(
		ctx,
		&user,
		"SELECT id, wallet_id, kyc_status, created_at, updated_at from machnet_users WHERE wallet_id = $1;",
		walletID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w, %s", machnet.ErrInternal, err)
	}

	return &user, nil
}

func GetUserByID(ctx context.Context, b Backends, id string) (*machnet.User, error) {
	var user machnet.User
	err := b.DB().GetContext(
		ctx,
		&user,
		"SELECT id, wallet_id, kyc_status, created_at, updated_at from machnet_users WHERE id = $1;",
		id,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w, %s", machnet.ErrInternal, err)
	}

	return &user, nil
}

func GetWidgetToken(ctx context.Context, b Backends, walletID string) (*machnet.WidgetToken, error) {
	user, err := GetUserByWalletID(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	token, err := b.External().GetFundingAccountWidgetToken(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &machnet.WidgetToken{
		Value:            token.Token,
		ExpiresInMinutes: token.ExpiryMinutes,
		UserID:           token.UserID,
	}, nil
}

func CreateReceiveBankAccount(ctx context.Context, b Backends, args machnet.CreateReceiveBankAccountArgs) (*machnet.ReceiveBankAccount, error) {
	// check if this account exists
	var ra machnet.ReceiveBankAccount
	err := b.DB().GetContext(
		ctx,
		&ra,
		`
		 	SELECT id, wallet_id, account_number, account_type, bank_id, branch_id, created_at, updated_at FROM machnet_receive_bank_accounts WHERE
		 	wallet_id=$1 AND account_number=$2 AND bank_id=$3 AND branch_id=$4 AND account_type=$5;
		`,
		args.WalletID,
		args.AccountNumber,
		args.BankID,
		args.BranchID,
		args.AccountType,
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	// it didn't exist so create it
	if errors.Is(err, sql.ErrNoRows) {
		insert := db.NewInsert("machnet_receive_bank_accounts").
			Value("wallet_id", args.WalletID).
			Value("account_number", args.AccountNumber).
			Value("account_type", args.AccountType).
			Value("bank_id", args.BankID).
			Value("branch_id", args.BranchID).
			Returning("id, wallet_id, account_number, account_type, bank_id, branch_id, created_at, updated_at")

		statement, values, err := insert.GetStatement()
		if err != nil {
			return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
		}

		err = b.DB().GetContext(ctx, &ra, statement, values...)
		if err != nil {
			return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
		}
	}

	// now check if linked account exists
	_, err = b.LinkedAccounts().GetByProviderID(ctx, linkedaccounts.GetByProviderIDArgs{
		Provider:   machnet.ProviderName,
		ProviderID: ra.ID,
		Type:       machnet.TypeReceiveBankAccount,
		WalletID:   args.WalletID,
	})
	if err != nil && !errors.Is(err, linkedaccounts.ErrNotFound) {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	// nope, didn't exist so create it
	if errors.Is(err, linkedaccounts.ErrNotFound) {
		mask := ra.AccountNumber
		if len(mask) > 4 {
			mask = mask[len(mask)-4:]
		}
		_, err = b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
			WalletID:   args.WalletID,
			Name:       args.Name,
			Mask:       mask,
			Provider:   machnet.ProviderName,
			ProviderID: ra.ID,
			Type:       machnet.TypeReceiveBankAccount,
		})
		if err != nil {
			return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
		}
	}

	return &ra, nil
}

func GetReceiveBankAccount(ctx context.Context, b Backends, id string) (*machnet.ReceiveBankAccount, error) {
	var ra machnet.ReceiveBankAccount
	err := b.DB().GetContext(
		ctx,
		&ra,
		"SELECT id, wallet_id, account_number, account_type, bank_id, branch_id, created_at, updated_at FROM machnet_receive_bank_accounts WHERE id=$1;",
		id,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w failed to find receivebank account (%s)", machnet.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &ra, nil
}

func CreateReceiveUser(ctx context.Context, b Backends, args machnet.CreateReceiveUserArgs) (*machnet.ReceiveUser, error) {
	insert := db.NewInsert("machnet_receive_users").
		Value("id", args.ExternalID).
		Value("send_user_id", args.SendUserID).
		Value("receive_wallet_id", args.ReceiveWalletID).
		Returning("id, send_user_id, receive_wallet_id, created_at, updated_at")

	statement, values, err := insert.GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	var ru machnet.ReceiveUser
	err = b.DB().GetContext(ctx, &ru, statement, values...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &ru, nil
}

func GetReceiveUser(ctx context.Context, b Backends, args machnet.GetReceiveUserArgs) (*machnet.ReceiveUser, error) {
	var ru machnet.ReceiveUser
	err := b.DB().GetContext(
		ctx,
		&ru,
		"SELECT id, send_user_id, receive_wallet_id, created_at, updated_at FROM machnet_receive_users WHERE receive_wallet_id=$1 and send_user_id=$2;",
		args.ReceiveWalletID,
		args.SendUserID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &ru, nil
}

func CreateReceiveUserBankAccount(ctx context.Context, b Backends, args machnet.CreateReceiveUserBankAccountArgs) (*machnet.ReceiveUserBankAccount, error) {
	insert := db.NewInsert("machnet_receive_user_bank_accounts").
		Value("id", args.ExternalID).
		Value("receive_user_id", args.ReceiveUserID).
		Value("receive_bank_account_id", args.ReceiveBankAccountID).
		Returning("id, receive_user_id, receive_bank_account_id, created_at, updated_at")

	statement, values, err := insert.GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	var rua machnet.ReceiveUserBankAccount
	err = b.DB().GetContext(ctx, &rua, statement, values...)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &rua, nil
}

func GetReceiveUserBankAccount(ctx context.Context, b Backends, args machnet.GetReceiveUserBankAccountArgs) (*machnet.ReceiveUserBankAccount, error) {
	var rua machnet.ReceiveUserBankAccount
	err := b.DB().GetContext(
		ctx,
		&rua,
		"SELECT id, receive_user_id, receive_bank_account_id, created_at, updated_at FROM machnet_receive_user_bank_accounts WHERE receive_bank_account_id=$1 AND receive_user_id=$2;",
		args.ReceiveBankAccountID,
		args.ReceiveUserID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &rua, nil
}

func GetBanks(ctx context.Context, b Backends, countryCode string) ([]machnet.Bank, error) {
	externalList, err := b.External().GetBanks(ctx, countryCode)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	banks := make([]machnet.Bank, len(externalList))
	for i, externalBank := range externalList {
		branches := make([]machnet.Branch, len(externalBank.Branches))
		for j, externalBranch := range externalBank.Branches {
			branches[j] = machnet.Branch{
				ID:   externalBranch.ID,
				Name: externalBranch.Name,
			}
		}
		banks[i] = machnet.Bank{
			ID:                        externalBank.ID,
			Name:                      externalBank.Name,
			Branches:                  branches,
			Country:                   externalBank.Country,
			TransactionSupportedTypes: externalBank.TransactionSupportedTypes,
			ReceivingCurrency:         externalBank.ReceivingCurrency,
		}
	}

	return banks, nil
}

func CreateUserWorkflowRef(ctx context.Context, b Backends, args machnet.CreateUserWorkflowRefArgs) (*machnet.UserWorkflowRef, error) {
	insert := db.NewInsert("machnet_users_workflow_ref").
		Value("user_id", args.UserID).Returning("user_id").
		Value("workflow_id", args.WorkflowID).Returning("workflow_id").
		Value("workflow_run_id", args.WorkflowRunID).Returning("workflow_run_id").
		Value("activity_name", args.ActivityName).Returning("activity_name")

	statement, values, err := insert.GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	var ref machnet.UserWorkflowRef
	err = b.DB().GetContext(
		ctx,
		&ref,
		statement,
		values...,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &ref, nil
}

func ListActiveUserWorkflowRefs(ctx context.Context, b Backends, userID string) ([]machnet.UserWorkflowRef, error) {
	var res []machnet.UserWorkflowRef
	err := b.DB().SelectContext(ctx, &res, "SELECT user_id, workflow_id, workflow_run_id, activity_name FROM  machnet_users_workflow_ref WHERE user_id=$1 AND completed=false", userID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return res, nil
}

func CompleteUserWorkflowRef(ctx context.Context, b Backends, args machnet.CreateUserWorkflowRefArgs) error {
	_, err := b.DB().ExecContext(ctx, "UPDATE machnet_users_workflow_ref SET completed=true, updated_at=now() WHERE completed=false AND user_id=$1 AND workflow_id=$2 AND workflow_run_id=$3 AND activity_name=$4",
		args.UserID, args.WorkflowID, args.WorkflowID, args.ActivityName)
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return nil
}

func CreateTransactionWorkflowRef(ctx context.Context, b Backends, args machnet.CreateTransactionWorkflowRefArgs) (*machnet.TransactionWorkflowRef, error) {
	insert := db.NewInsert("machnet_transactions_workflow_ref").
		Value("id", args.ID).Returning("id").
		Value("send_user_id", args.SendUserID).Returning("send_user_id").
		Value("workflow_id", args.WorkflowID).Returning("workflow_id").
		Value("workflow_run_id", args.WorkflowRunID).Returning("workflow_run_id").
		Value("activity_name", args.ActivityName).Returning("activity_name")

	statement, values, err := insert.GetStatement()
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	var ref machnet.TransactionWorkflowRef
	err = b.DB().GetContext(
		ctx,
		&ref,
		statement,
		values...,
	)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &ref, nil
}

func GetTransactionWorkflowRef(ctx context.Context, b Backends, userID string, transactionID string) (*machnet.TransactionWorkflowRef, error) {
	var ref machnet.TransactionWorkflowRef
	err := b.DB().GetContext(
		ctx,
		&ref,
		"SELECT id, send_user_id, workflow_id, workflow_run_id, activity_name FROM machnet_transactions_workflow_ref WHERE id=$1 AND send_user_id=$2;",
		transactionID,
		userID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &ref, nil
}

func GetWorkflowRef(ctx context.Context, b Backends, workflowID, activityName string) (*machnet.TransactionWorkflowRef, error) {
	var ref machnet.TransactionWorkflowRef
	err := b.DB().GetContext(
		ctx,
		&ref,
		"SELECT id, send_user_id, workflow_id, workflow_run_id, activity_name FROM machnet_transactions_workflow_ref WHERE workflow_id=$1 AND activity_name=$2;",
		workflowID,
		activityName,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &ref, nil
}

func CreateWallet(ctx context.Context, b Backends, args machnet.CreateWalletArgs) (*linkedaccounts.LinkedAccount, error) {
	var existingIDAndNickname []dbWallet
	err := b.DB().SelectContext(ctx, &existingIDAndNickname, "SELECT id, nickname FROM machnet_wallets WHERE send_user_id=$1;", args.SendUserID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	if len(existingIDAndNickname) == 1 {
		if !strings.EqualFold(existingIDAndNickname[0].Nickname, strings.TrimSpace(args.Nickname)) {
			return nil, fmt.Errorf("%w SendUserId=%s already has a wallet.", machnet.ErrUserHasExistingWallet, args.SendUserID)
		}

		sendUser, err := GetUserByID(ctx, b, args.SendUserID)
		if err != nil {
			return nil, err
		}

		return b.LinkedAccounts().GetByProviderID(ctx, linkedaccounts.GetByProviderIDArgs{
			Provider:   machnet.ProviderName,
			ProviderID: existingIDAndNickname[0].ID,
			Type:       machnet.TypeWallet,
			WalletID:   sendUser.WalletID,
		})
	}

	if len(existingIDAndNickname) > 1 {
		return nil, fmt.Errorf("%w SendUserId=%s has more than 1 wallet.", machnet.ErrInternal, args.SendUserID)
	}

	_, linkedAccount, err := createWallet(ctx, b, args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return linkedAccount, nil
}

func createWallet(ctx context.Context, b Backends, args machnet.CreateWalletArgs) (*dbWallet, *linkedaccounts.LinkedAccount, error) {
	sendUser, err := GetUserByID(ctx, b, args.SendUserID)
	if err != nil {
		return nil, nil, err
	}

	externalWallet, err := b.External().CreateUserWallet(ctx, args.SendUserID, args.Nickname)
	if err != nil {
		return nil, nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	insert := db.NewInsert("machnet_wallets").
		Value("id", externalWallet.ID).Returning("id").
		Value("nickname", strings.TrimSpace(externalWallet.NickName)).Returning("nickname").
		Value("send_user_id", externalWallet.UserID).Returning("send_user_id").
		Returning("created_at").Returning("updated_at")

	sql, values, err := insert.GetStatement()
	if err != nil {
		return nil, nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	var dbWallet dbWallet
	err = b.DB().GetContext(ctx, &dbWallet, sql, values...)
	if err != nil {
		return nil, nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	la, err := b.LinkedAccounts().Create(ctx, &linkedaccounts.CreateArgs{
		WalletID:   sendUser.WalletID,
		Name:       args.Nickname,
		Provider:   machnet.ProviderName,
		ProviderID: externalWallet.ID,
		Type:       machnet.TypeWallet,
		Mask:       "Fynbos Cash",
	})
	if err != nil {
		return nil, nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &dbWallet, la, nil
}

func GetWallet(ctx context.Context, b Backends, id string) (*machnet.Wallet, error) {
	var wallet dbWallet
	err := b.DB().GetContext(ctx, &wallet, "SELECT id, send_user_id, nickname, created_at, updated_at FROM machnet_wallets WHERE id=$1;", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w wallet not found. id=%s", machnet.ErrNotFound, id)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	externalWallet, err := b.External().GetUserWallet(ctx, wallet.SendUserID, wallet.ID)
	if errors.Is(err, external.ErrNotFound) {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &machnet.Wallet{
		ID:               wallet.ID,
		SendUserID:       wallet.SendUserID,
		Nickname:         wallet.Nickname,
		AvailableBalance: uint64(externalWallet.Balance.AvailableBalance * float64(100)),
		Balance:          uint64(externalWallet.Balance.Balance * float64(100)),
		CreatedAt:        wallet.CreatedAt,
	}, nil
}

func GetWalletByWalletID(ctx context.Context, b Backends, walletID string) (*machnet.Wallet, error) {
	user, err := GetUserByWalletID(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	var wallet dbWallet
	err = b.DB().GetContext(ctx, &wallet, "SELECT id, send_user_id, nickname, created_at, updated_at FROM machnet_wallets WHERE send_user_id=$1;", user.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w wallet not found. fynbos wallet_id=%s", machnet.ErrNotFound, walletID)
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	externalWallet, err := b.External().GetUserWallet(ctx, wallet.SendUserID, wallet.ID)
	if errors.Is(err, external.ErrNotFound) {
		return nil, machnet.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return &machnet.Wallet{
		ID:               wallet.ID,
		SendUserID:       wallet.SendUserID,
		Nickname:         wallet.Nickname,
		AvailableBalance: uint64(externalWallet.Balance.AvailableBalance * float64(100)),
		Balance:          uint64(externalWallet.Balance.Balance * float64(100)),
		CreatedAt:        wallet.CreatedAt,
	}, nil
}

func SetKYCInProgress(ctx context.Context, b Backends, userID string) error {

	_, err := b.DB().ExecContext(ctx, "UPDATE machnet_users SET updated_at=now(), kyc_status=$1 WHERE id=$2 AND kyc_status=$3", machnet.KYCStatusInProgress, userID, machnet.KYCStatusRetry)
	if err != nil {
		return fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return nil
}

func GetStatement(ctx context.Context, b Backends, walletID, period string) ([]byte, error) {
	wallet, err := GetWallet(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	machnetUser, err := GetUserByID(ctx, b, wallet.SendUserID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	userDetails, err := b.KYC().GetIndividualDetails(ctx, machnetUser.WalletID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	periodStart, err := time.Parse("2006-01", period)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	trxs, err := b.Transactions().ListTransactionsInRange(
		ctx,
		machnetUser.WalletID, // fynbos wallet
		transactions.TransactionRangeFilter{
			StartTimestamp: periodStart,                                    // 1st day of the month at 00:00:00
			EndTimestamp:   periodStart.AddDate(0, 1, 0).Add(-time.Second), // last day of the month at 23:59:59
		})
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	var statementRows []statements.TransactionTableRow
	for _, trx := range trxs {
		if trx.State != transactions.StateCompleted {
			continue
		}
		statementRows = append(statementRows, mapToStatementRows(trx)...)
	}

	statementBalance := currency.Amount{
		Value:    wallet.AvailableBalance,
		Currency: "USD",
	}
	pdf, err := b.Statements().GenerateWalletStatementPDF(ctx, statements.GenerateWalletStatementArgs{
		Name:         fmt.Sprintf("%s %s", userDetails.FirstName, userDetails.LastName),
		Period:       fmt.Sprintf("%s-%s", periodStart.Format("02 Jan 2006"), periodStart.AddDate(0, 1, -1).Format("02 Jan 2006")),
		BalanceDate:  time.Now().Format("02 Jan 2006"),
		Balance:      statementBalance.Format(),
		Transactions: statementRows,
		AccountID:    wallet.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	return pdf, nil
}

func mapToStatementRows(trx transactions.Transaction) []statements.TransactionTableRow {
	var ret []statements.TransactionTableRow
	switch trx.Type {
	case transactions.TransactionTypeMachnetWalletWithdrawal:
		ret = append(ret, mapWalletWithdrawalToStatementRow(trx)...)
	case transactions.TransactionTypeMachnetWalletTopUp:
		ret = append(ret, mapWalletTopupToStatementRow(trx)...)
	case transactions.TransactionTypeOpenOutgoingPayment:
		ret = append(ret, mapOutgoingPaymentToStatementRow(trx)...)
	case transactions.TransactionTypeOpenPaymentIncoming:
		ret = append(ret, mapIncomingPaymentToStatementRow(trx)...)
	}

	return ret
}

func mapIncomingPaymentToStatementRow(trx transactions.Transaction) []statements.TransactionTableRow {
	var ret []statements.TransactionTableRow
	for _, transfer := range trx.Transfers {
		var description string
		switch transfer.Type {
		case transactions.TransferTypeCreditMachnetWallet:
			description = fmt.Sprintf("Payment from %s", trx.Source)
		default:
			continue
		}

		ret = append(ret, statements.TransactionTableRow{
			Date:        transfer.Timestamp.Format("02 Jan 2006"),
			Description: description,
			Amount:      transfer.Amount.Format(),
			RecieptID:   trx.ID,
		})
	}

	return ret
}

func mapOutgoingPaymentToStatementRow(trx transactions.Transaction) []statements.TransactionTableRow {
	var ret []statements.TransactionTableRow
	for _, transfer := range trx.Transfers {
		var description string
		switch transfer.Type {
		case transactions.TransferTypeDebitMachnetWallet:
			description = fmt.Sprintf("Payment to %s", trx.Destination)
		default:
			continue
		}

		ret = append(ret, statements.TransactionTableRow{
			Date:        transfer.Timestamp.Format("02 Jan 2006"),
			Description: description,
			Amount:      fmt.Sprintf("-%s", transfer.Amount.Format()),
			RecieptID:   trx.ID,
		})
	}

	return ret
}

func mapWalletTopupToStatementRow(trx transactions.Transaction) []statements.TransactionTableRow {
	var ret []statements.TransactionTableRow
	for _, transfer := range trx.Transfers {
		var description string
		switch transfer.Type {
		case transactions.TransferTypeCreditMachnetWallet:
			description = "Top up cash balance"
		default:
			continue
		}

		ret = append(ret, statements.TransactionTableRow{
			Date:        transfer.Timestamp.Format("02 Jan 2006"),
			Description: description,
			Amount:      transfer.Amount.Format(),
			RecieptID:   trx.ID,
		})
	}

	return ret
}

func mapWalletWithdrawalToStatementRow(trx transactions.Transaction) []statements.TransactionTableRow {
	var ret []statements.TransactionTableRow
	for _, transfer := range trx.Transfers {
		var description string
		switch transfer.Type {
		case transactions.TransferTypeDebitMachnetWallet:
			description = "Withdrawal from cash balance"
		default:
			continue
		}

		ret = append(ret, statements.TransactionTableRow{
			Date:        transfer.Timestamp.Format("02 Jan 2006"),
			Description: description,
			Amount:      fmt.Sprintf("-%s", transfer.Amount.Format()),
			RecieptID:   trx.ID,
		})
	}

	return ret
}

func GetUserLimits(ctx context.Context, b Backends, walletID string) (*machnet.UserLimits, error) {
	user, err := GetUserByWalletID(ctx, b, walletID)
	if err != nil {
		return nil, err
	}

	ul, err := b.External().GetUserLimits(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("%w %s", machnet.ErrInternal, err)
	}

	var resp machnet.UserLimits
	for _, l := range ul {
		remaining := machnet.RemainingLimit{
			Annual:     currency.FromFloat64(l.Remaining.Annual, currency.USD),
			Daily:      currency.FromFloat64(l.Remaining.Daily, currency.USD),
			Monthly:    currency.FromFloat64(l.Remaining.Monthly, currency.USD),
			WalletHold: currency.FromFloat64(l.Remaining.WalletHold, currency.USD),
		}
		if l.Type == external.LimitTypeLoad {
			resp.FundWallet = remaining
		}
		if l.Type == external.LimitTypeWithdraw {
			resp.Withdrawal = remaining
		}
		if l.Type == external.LimitTypeTransfer {
			resp.Transfer = remaining
		}
	}

	return &resp, nil
}
