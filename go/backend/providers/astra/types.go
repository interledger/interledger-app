package astra

import (
	"context"
	"os"
	"sync"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/env"
)

const (
	ProviderName = "astra"
	TypeCard     = "card"
)

var (
	accountNumber string
	routingNumber string

	once = sync.Once{}
)

func AccountNumber() string {
	once.Do(func() {
		if env.IsProd() {
			accountNumber = os.Getenv("ASTRA_ACCOUNT_NUMBER")
			routingNumber = os.Getenv("ASTRA_ROUTING_NUMBER")
		} else {
			accountNumber = "90643128"
			routingNumber = "434253106"
		}
	})

	return accountNumber
}

func RoutingNumber() string {
	once.Do(func() {
		if env.IsProd() {
			accountNumber = os.Getenv("ASTRA_ACCOUNT_NUMBER")
			routingNumber = os.Getenv("ASTRA_ROUTING_NUMBER")
		} else {
			accountNumber = "90643128"
			routingNumber = "434253106"
		}
	})

	return routingNumber
}

type TransferStatus string

const (
	TransferStatusPending   = "pending"
	TransferStatusProcessed = "processed"
	TransferStatusCancelled = "cancelled"
	TransferStatusFailed    = "failed"

	RoutineStatusRequiresUserVerification = "requires_user_verification"
	RoutineStatusPendingAccountAuth       = "pending_account_authorization"
	RoutineStatusUserSuspended            = "user_suspended"
	RoutineStatusActive                   = "active"
	RoutineStatusInactive                 = "inactive"
	RoutineStatusCancelled                = "cancelled"
	RoutineStatusFailed                   = "failed"
	RoutineStatusCompleted                = "completed"
)

type Await func(ctx context.Context, result interface{}) error

type CreateCardArgs struct {
	WalletID           string
	BasisTheoryTokenID string
}

type CardToAccountArgs struct {
	WalletID            string
	IdempotencyKey      string
	Name                string
	Amount              currency.Amount
	ClientCorrelationID string `validate:"len=8"` // Exactly 8 Chars
	DebitFeePercent     int
	CardID              string
}

type AccountToCardsArgs struct {
	WalletID        string
	IdempotencyKey  string
	Name            string
	Amount          currency.Amount
	DebitFeePercent int
	CardID          string
}

type Transfer struct {
	ID                    string
	RoutineType           string
	RoutineName           string
	RoutineID             string
	ClientCorrelationID   string
	Amount                currency.Amount
	PaymentType           string
	AstraSettlementReason string
	FailureReason         string
	Status                string
}

type Routine struct {
	ID        string
	Active    bool
	Blocked   bool
	Status    string
	Name      string
	Type      string
	StartDate string
}
