package noop

//go:generate mockgen -destination=./mock.go -package=noop -source=./service.go

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	pacioliv1 "gitlab.com/fynbos/proto/pacioli/v1"
)

var (
	ErrInternal = errors.New("noop: internal error.")
)

const (
	Verified  string = "verified"
	Retry     string = "retry"
	Kba       string = "kba"
	Document  string = "document"
	Suspended string = "suspended"
)

type Service interface {
	VerifyBankAccount(ctx context.Context, args *VerifyArgs) error
	InitiateBankDeposit(ctx context.Context, args *BankDepositArgs) error
	InitiateBankWithdrawal(ctx context.Context, args *BankWithdrawalArgs) error
	InitiateOutgoingPayment(ctx context.Context, args *OutgoingPaymentArgs) error

	GetEquityAccountID() string
	GetLedgerID() uint32
	CreateCustomer(args *CreateCustomerArgs) (*Customer, error)
}

type service struct {
	validator *validator.Validate
	// TODO: configuring and management of ledgers and ledger accounts that this provider
	// acts upon.
	ledgerID        uint32
	equityAccountID string
	pacioliClient   pacioliv1.PacioliServiceClient
}

type ServiceArgs struct {
	// TODO: configuring and management of ledgers and ledger accounts
	LedgerID      uint32
	EquityAccID   string `validate:"required"`
	PacioliTenant string
	PacioliClient pacioliv1.PacioliServiceClient `validate:"required"`
}

func NewService(args ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, err
	}

	return &service{
		validator:       validator,
		ledgerID:        args.LedgerID,
		equityAccountID: args.EquityAccID,
		pacioliClient:   args.PacioliClient,
	}, nil
}

type NoopBankAccount struct {
	ID                 string
	Name               string
	Mask               string
	VerificationStatus string
	Type               string
}

type VerifyArgs struct {
	// place holder
}

func (s *service) VerifyBankAccount(ctx context.Context, args *VerifyArgs) error {
	fmt.Println("Noop verifying bank account.")

	return nil
}

type BankDepositArgs struct {
	IdentityID      string `validate:"required,uuid4"`
	FundingSourceID string `validate:"required,uuid4"`
	Amount          uint64 `validate:"gt=0"`
}

func (s *service) InitiateBankDeposit(ctx context.Context, args *BankDepositArgs) error {
	fmt.Println("Iniating bank deposit", args.Amount)
	return nil
}

type BankWithdrawalArgs struct {
	Amount uint64 `validate:"required,gt=0"` // Min amount can be changed according to provider
}

func (s *service) InitiateBankWithdrawal(ctx context.Context, args *BankWithdrawalArgs) error {
	fmt.Println("Initiating bank withdrawal", args.Amount)
	return nil
}

type OutgoingPaymentArgs struct {
	Amount uint64 `validate:"required,gt=0"`
	To     string `validate:"required"`
}

func (s *service) InitiateOutgoingPayment(ctx context.Context, args *OutgoingPaymentArgs) error {
	fmt.Printf("Initiating outgoing payment. to: %s amount: %d.\n", args.To, args.Amount)
	return nil
}

func (s *service) GetEquityAccountID() string {
	return s.equityAccountID
}

func (s *service) GetLedgerID() uint32 {
	return s.ledgerID
}

type CreateCustomerArgs struct {
	FirstName   string
	LastName    string
	Email       string
	Address1    string
	Address2    string
	State       string
	City        string
	PostalCode  string
	DateOfBirth string
	Ssn         string
}

type Customer struct {
	CreateCustomerArgs
	ID      string
	Status  string
	Created string
	Links   map[string]map[string]string
}

func (self *service) CreateCustomer(args *CreateCustomerArgs) (*Customer, error) {
	return &Customer{
		ID:                 uuid.NewString(),
		CreateCustomerArgs: *args,
		Created:            time.Now().String(),
		Status:             Verified,
	}, nil
}
