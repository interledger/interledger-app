package machnet

import (
	"context"
)

const (
	ProviderName           = "machnet"
	TypeReceiveBankAccount = "receiveBankAccount"
	TypeSendCard           = "sendCard"
	TypeWallet             = "wallet"
)

type Await func(context.Context, interface{}) error

type User struct {
	ID        string    `db:"id"`
	WalletID  string    `db:"wallet_id"`
	CreatedAt string    `db:"created_at"`
	UpdatedAt string    `db:"updated_at"`
	KYCStatus KYCStatus `db:"kyc_status"`
}

type CreateArgs struct {
	WalletID   string
	ExternalID string
}

type WidgetToken struct {
	Value            string
	ExpiresInMinutes int
	UserID           string
}

type CreateTransactionArgs struct {
	FromLinkedAccountID string
	ToLinkedAccountID   string
	Amount              float64
	Currency            string
	IPAddress           string
}

type (
	// This will underpin a linked account and is used to create Machnet receive user accounts.
	ReceiveBankAccount struct {
		ID            string          `db:"id"`
		WalletID      string          `db:"wallet_id"`
		AccountNumber string          `db:"account_number"`
		AccountType   BankAccountType `db:"account_type"`
		BankID        uint32          `db:"bank_id"`
		BranchID      uint32          `db:"branch_id"`
		CreatedAt     string          `db:"created_at"`
		UpdatedAt     string          `db:"updated_at"`
	}

	CreateReceiveBankAccountArgs struct {
		WalletID      string
		AccountNumber string
		AccountType   BankAccountType
		BankID        uint32
		BranchID      uint32
		Name          string
	}
)

type BankAccountType int

const (
	BankAccountTypeUnknown BankAccountType = 0
	BankAccountTypeSavings BankAccountType = 1
	BankAccountTypeCheque  BankAccountType = 2
)

type (
	ReceiveUser struct {
		ID              string `db:"id"`
		SendUserID      string `db:"send_user_id"`
		ReceiveWalletID string `db:"receive_wallet_id"`
		CreatedAt       string `db:"created_at"`
		UpdatedAt       string `db:"updated_at"`
	}

	CreateReceiveUserArgs struct {
		ExternalID      string
		SendUserID      string
		ReceiveWalletID string
	}

	GetReceiveUserArgs struct {
		ReceiveWalletID string
		SendUserID      string
	}
)

type (
	ReceiveUserBankAccount struct {
		ID                   string `db:"id"`
		ReceiveUserID        string `db:"receive_user_id"`
		ReceiveBankAccountID string `db:"receive_bank_account_id"`
		CreatedAt            string `db:"created_at"`
		UpdatedAt            string `db:"updated_at"`
	}

	CreateReceiveUserBankAccountArgs struct {
		ExternalID           string
		ReceiveUserID        string
		ReceiveBankAccountID string
	}

	GetReceiveUserBankAccountArgs struct {
		ReceiveUserID        string
		ReceiveBankAccountID string
	}
)

type KYCStatus int

const (
	KYCStatusUnknown       KYCStatus = 0
	KYCStatusInProgress    KYCStatus = 1 // User's KYC is in progress.
	KYCStatusVerified      KYCStatus = 2 // User’s KYC is complete and user details are verified.
	KYCStatusRetry         KYCStatus = 3 // If user KYC is not complete and additional details have to be provided. Once the details have been collected, ‘Initiate KYC’ API needs to be called again.
	KYCStatusSuspended     KYCStatus = 4 // When a user's KYC is rejected during KYC process.
	KYCStatusReviewPending KYCStatus = 5 // When a user’s KYC process is in review state.
)

type Branch struct {
	ID   uint32
	Name string
}

type Bank struct {
	ID                        uint32
	Name                      string
	Branches                  []Branch
	Country                   string
	TransactionSupportedTypes []string
	ReceivingCurrency         []string
}

type (
	TransactionWorkflowRef struct {
		ID            string
		SendUserID    string `db:"send_user_id"`
		WorkflowID    string `db:"workflow_id"`
		WorkflowRunID string `db:"workflow_run_id"`
		ActivityName  string `db:"activity_name"`
	}

	CreateTransactionWorkflowRefArgs struct {
		ID            string
		SendUserID    string
		WorkflowID    string
		WorkflowRunID string
		ActivityName  string
	}

	CreateUserWorkflowRefArgs struct {
		UserID        string
		WorkflowID    string
		WorkflowRunID string
		ActivityName  string
	}

	UserWorkflowRef struct {
		UserID        string `db:"user_id"`
		WorkflowID    string `db:"workflow_id"`
		WorkflowRunID string `db:"workflow_run_id"`
		ActivityName  string `db:"activity_name"`
	}
)

type (
	Wallet struct {
		ID               string
		SendUserID       string
		Nickname         string
		AvailableBalance uint64
		Balance          uint64
	}

	CreateWalletArgs struct {
		Nickname   string
		SendUserID string
	}

	WithdrawFromWalletArgs struct {
		Amount                uint64 `validate:"gt=0"`
		WalletLinkedAccountID string `validate:"required,uuid"`
		ToLinkedAccountID     string `validate:"required,uuid"`
		IpAddress             string `validate:"ip_addr"`
	}

	WalletWithdrawal struct {
		ID                string
		Amount            uint64
		ToLinkedAccountID string
		Status            string
	}
)
