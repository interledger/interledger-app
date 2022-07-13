package ledger

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	tigerbeetle_go "github.com/coilhq/tigerbeetle-go"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	tigerbeetleTypes "github.com/coilhq/tigerbeetle-go/pkg/types"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var (
	ErrInvalidArg = errors.New("pacioli: invalid argument.")
	ErrNotFound   = errors.New("pacioli: not found.")
	ErrDuplicate  = errors.New("pacioli: duplicate.")
	ErrInternal   = errors.New("pacioli: internal error.")
)

const (
	LEDGER_OK                          uint8 = 0
	LEDGER_EXISTS_WITH_DIFFERENT_NAME  uint8 = 1
	LEDGER_EXISTS_WITH_DIFFERENT_ASSET uint8 = 2
	LEDGER_EXISTS_WITH_DIFFERENT_SCALE uint8 = 3

	ACCOUNT_LEDGER_DOES_NOT_EXIST uint8 = 0 // TB account errors start at 1
)

// Models
type Ledger struct {
	ID        uint16 `json:"id"` // maps to TigerBeetle's `unit` (soon to be ledger) field on an account.
	Name      string `json:"name"`
	Asset     string `json:"string"`
	Scale     uint8  `json:"scale"`
	CreatedAt string `db:"created_at"`
	UpdatedAt string `db:"updated_at"`
}

type Account struct {
	ID              string
	LedgerID        uint16
	Flags           AccountFlags
	Code            uint16
	DebitsReserved  uint64
	DebitsAccepted  uint64
	CreditsReserved uint64
	CreditsAccepted uint64
}

type Transfer struct {
	ID              string
	LedgerID        uint16 // this field is coming soon to a TigerBeetle near you.
	DebitAccountID  string
	CreditAccountID string
	Amount          uint64
	Flags           TransferFlags
	Code            uint32
	Timeout         uint64
}

type CommitFlags = tigerbeetleTypes.CommitFlags
type TransferFlags = tigerbeetleTypes.TransferFlags
type AccountFlags = tigerbeetleTypes.AccountFlags
type EventResult = tigerbeetleTypes.EventResult

type Service interface {

	// This is declaritive and will not fail if the ledger exists. It will fail if one exists with
	// different fields.
	ConfigureLedgers(ctx context.Context, args []ConfigureLedgerArgs) ([]EventResult, error)
	GetLedgers(ctx context.Context, ledgerIDs []uint32) ([]Ledger, error)

	// This is declaritive and will not fail if the account exists. It will fail if one exists with
	// different fields.
	ConfigureAccounts(ctx context.Context, args []ConfigureAccountArgs) ([]EventResult, error)
	GetAccounts(ctx context.Context, accountIDs []string) ([]Account, error)
	CreateTransfers(ctx context.Context, args []CreateTransferArgs) ([]EventResult, error)
	GetTransfers(ctx context.Context, transferIDs []string) ([]Transfer, error)
	CommitTransfers(ctx context.Context, args []CommitTransferArgs) ([]EventResult, error)
}

type service struct {
	db        *sqlx.DB
	tb        tigerbeetle_go.Client
	validator *validator.Validate
}

type ServiceArgs struct {
	Db *sqlx.DB              `validate:"required"`
	Tb tigerbeetle_go.Client `validate:"required"`
}

func NewService(args *ServiceArgs) (Service, error) {
	validator := validator.New()
	err := validator.Struct(args)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInvalidArg)
	}

	return &service{db: args.Db, tb: args.Tb, validator: validator}, nil
}

type ConfigureLedgerArgs struct {
	ID    uint16
	Name  string `validate:"required"`
	Asset string `validate:"required"`
	Scale uint8  `validate:"gt=0"`
}

// This is declaritive and will not fail if the ledger exists. It will fail if one exists with
// different fields.
func (s *service) ConfigureLedgers(
	ctx context.Context,
	args []ConfigureLedgerArgs,
) ([]EventResult, error) {
	ledgerIds := make([]uint32, len(args))
	for i, ledger := range args {
		err := s.validator.Struct(ledger)
		if err != nil {
			return nil, fmt.Errorf("%s %d %s %w", "index: ", i, err.Error(), ErrInvalidArg)
		}

		ledgerIds[i] = uint32(ledger.ID)
	}

	existingLedgers, err := s.GetLedgers(ctx, ledgerIds)
	if err != nil {
		return nil, err
	}

	errorEvents := []EventResult{}
	createdLedgers := []Ledger{}
	err = crdbsqlx.ExecuteTx(ctx, s.db, nil, func(tx *sqlx.Tx) error {
		for i, ledger := range args {
			exists := false
			for _, existing := range existingLedgers {
				if ledger.ID == existing.ID {
					exists = true
					result := canCreateLedger(ledger, existing)
					if result != LEDGER_OK {
						errorEvents = append(errorEvents, EventResult{
							Index: uint32(i),
							Code:  uint32(result),
						})
						break
					}

					break
				}
			}

			// check for duplicates from the ones we just created
			for _, created := range createdLedgers {
				if ledger.ID == created.ID {
					exists = true
					result := canCreateLedger(ledger, created)
					if result != LEDGER_OK {
						errorEvents = append(errorEvents, EventResult{
							Index: uint32(i),
							Code:  uint32(result),
						})
						break
					}

					break
				}
			}

			if err != nil {
				return err
			}
			if !exists {
				_, err = tx.ExecContext(
					ctx,
					`INSERT INTO ledgers (id, name, asset, scale) VALUES ($1, $2, $3, $4);`,
					ledger.ID,
					ledger.Name,
					ledger.Asset,
					ledger.Scale,
				)
				if err != nil {
					return fmt.Errorf("%s %d %s %w", "index: ", i, err.Error(), ErrInternal)
				}

				createdLedgers = append(createdLedgers, Ledger{
					ID:    ledger.ID,
					Name:  ledger.Name,
					Asset: ledger.Asset,
					Scale: ledger.Scale,
				})
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return errorEvents, nil
}

func canCreateLedger(args ConfigureLedgerArgs, existingLedger Ledger) uint8 {
	if args.Name != existingLedger.Name {
		return LEDGER_EXISTS_WITH_DIFFERENT_NAME
	}

	if args.Asset != existingLedger.Asset {
		return LEDGER_EXISTS_WITH_DIFFERENT_ASSET
	}

	if args.Scale != existingLedger.Scale {
		return LEDGER_EXISTS_WITH_DIFFERENT_SCALE
	}

	return LEDGER_OK
}

func (s service) GetLedgers(ctx context.Context, ids []uint32) ([]Ledger, error) {
	// TODO: ACL

	var ledgers []Ledger
	query, args, err := sqlx.In("SELECT * FROM ledgers WHERE id IN (?);", ids)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInternal)
	}
	err = s.db.SelectContext(ctx, &ledgers, s.db.Rebind(query), args...)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInternal)
	}

	return ledgers, nil
}

type ConfigureAccountArgs struct {
	ID       string `validate:"required,uuid4"`
	LedgerID uint16
	Code     uint16
	Flags    AccountFlags
}

// Helper function to convert uuids into u128 needed for TigerBeetle IDs.
// TODO: see if there is a better way to do this.
func UuidToU128(value string) (*tigerbeetleTypes.Uint128, error) {
	src := strings.Replace(value, "-", "", -1)
	ret := new(tigerbeetleTypes.Uint128)
	bytesWritten, err := hex.Decode(ret[:], []byte(src))
	if err != nil {
		return nil, err
	}
	if bytesWritten > 16 {
		return nil, errors.New("String could not be converted into uint128.")
	}

	return ret, nil
}

// Helper function to extract the uuid we put into the u128.
func U128ToUuid(value tigerbeetleTypes.Uint128) string {
	s := hex.EncodeToString(value[:])
	ret := s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]

	return ret
}

// This is declaritive and will not fail if the account exists. It will fail if one exists with
// different fields.
func (s *service) ConfigureAccounts(
	ctx context.Context,
	args []ConfigureAccountArgs,
) ([]EventResult, error) {
	// TODO: ACL
	ledgerIDs := []uint32{}
	keys := map[uint16]uint8{}
	const (
		LOOKING_UP uint8 = 1
		EXISTS     uint8 = 2
	)
	// dedupe ledgerIDs by marking them as being looked up.
	for _, account := range args {
		if _, present := keys[account.LedgerID]; !present {
			keys[account.LedgerID] = LOOKING_UP
			ledgerIDs = append(ledgerIDs, uint32(account.LedgerID))
		}
	}

	ledgers, err := s.GetLedgers(ctx, ledgerIDs)
	if err != nil {
		return nil, err
	}

	// mark these ledgers as existing.
	for _, ledger := range ledgers {
		keys[ledger.ID] = EXISTS
	}

	// size to len(args) to avoid appending
	eventErrors := make([]EventResult, len(args))
	errors := uint32(0) // number of errors
	accountsToCreate := make([]tigerbeetleTypes.Account, len(args))
	index := uint32(0) // number of accounts to create

	// stores the mapping from the tb create account args to the original args
	mapToEventErrorSlot := map[uint32]uint32{}
	for i, acc := range args {
		err := s.validator.Struct(acc)
		if err != nil {
			return nil, fmt.Errorf("%s %d %s %w", "index: ", i, err.Error(), ErrInvalidArg)
		}

		if keys[acc.LedgerID] != EXISTS {
			eventErrors[errors] = EventResult{
				Index: uint32(i),
				Code:  uint32(ACCOUNT_LEDGER_DOES_NOT_EXIST),
			}
			errors++
			continue
		}

		tbAccID, err := UuidToU128(acc.ID)
		if err != nil {
			return nil, err
		}
		accountsToCreate[index] = tigerbeetleTypes.Account{
			ID:    *tbAccID,
			Unit:  acc.LedgerID,
			Code:  acc.Code,
			Flags: acc.Flags.ToUint32(),
		}
		mapToEventErrorSlot[index] = uint32(i)
		index++
	}
	tbEventErrors, err := s.tb.CreateAccounts(accountsToCreate[:index])
	// this error will be due to connection / io buffer issues
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), ErrInternal)
	}
	// map the tbEventErrors to our eventErrors
	for _, tbErr := range tbEventErrors {
		index, present := mapToEventErrorSlot[tbErr.Index]
		if !present {
			// the mapping is broken
			panic("Unable to map Tb event errors back to our create account errors.")
		}
		if tbErr.Code != tigerbeetleTypes.AccountExists {
			eventErrors[errors] = EventResult{Index: index, Code: tbErr.Code}
			errors++
		}
	}

	return eventErrors[:errors], nil
}

func (s *service) GetAccounts(
	ctx context.Context,
	accountIDs []string,
) ([]Account, error) {
	tbAccIDs := []tigerbeetleTypes.Uint128{}
	for _, id := range accountIDs {
		_, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("Account ID must be a uuid. %w", ErrInvalidArg)
		}
		accID, err := UuidToU128(id)
		if err != nil {
			return nil, err
		}

		tbAccIDs = append(tbAccIDs, *accID)
	}

	results, err := s.tb.LookupAccounts(tbAccIDs)
	if err != nil {
		return nil, err
	}

	ret := make([]Account, len(results))
	for i, result := range results {
		ret[i] = Account{
			ID:              U128ToUuid(result.ID),
			LedgerID:        result.Unit,
			Code:            result.Code,
			DebitsReserved:  result.DebitsReserved,
			DebitsAccepted:  result.DebitsAccepted,
			CreditsReserved: result.CreditsReserved,
			CreditsAccepted: result.CreditsAccepted,
			Flags: AccountFlags{
				Linked:                     result.Flags&(1<<0) == 1,
				DebitsMustNotExceedCredits: result.Flags&(1<<1) == 2,
				CreditsMustNotExceedDebits: result.Flags&(1<<2) == 4,
			},
		}
	}

	// TODO: ACL on accounts

	return ret, nil
}

type CreateTransferArgs struct {
	ID              string `validate:"required,uuid4"`
	Amount          uint64 `validate:"gt=0"`
	DebitAccountID  string `validate:"required,uuid4"`
	CreditAccountID string `validate:"required,uuid4"`
	Flags           TransferFlags
	Code            uint32
	Timeout         uint64
}

// TODO: Assuming that IDs are uuids and are being generated by another service for now. Might be better to be
// unopinionated and accept []byte.
func (s *service) CreateTransfers(ctx context.Context, args []CreateTransferArgs) ([]EventResult, error) {
	// TODO: collect ledgers and perform ACL. TigerBeetle will introduce the ledger field onto
	// a transfer.
	tbTransfers := make([]tigerbeetleTypes.Transfer, len(args))
	for i, transfer := range args {
		err := s.validator.Struct(transfer)
		if err != nil {
			return nil, fmt.Errorf("%s %w", err.Error(), ErrInvalidArg)
		}
		transferID, err := UuidToU128(transfer.ID)
		if err != nil {
			return nil, err
		}

		debitAccountID, err := UuidToU128(transfer.DebitAccountID)
		if err != nil {
			return nil, err
		}

		creditAccountID, err := UuidToU128(transfer.CreditAccountID)
		if err != nil {
			return nil, err
		}

		tbTransfers[i] = tigerbeetleTypes.Transfer{
			ID:              *transferID,
			DebitAccountID:  *debitAccountID,
			CreditAccountID: *creditAccountID,
			Amount:          transfer.Amount,
			Code:            transfer.Code,
			Flags:           transfer.Flags.ToUint32(),
			Timeout:         transfer.Timeout,
		}
	}

	eventErrors, err := s.tb.CreateTransfers(tbTransfers)
	if err != nil {
		return nil, err
	}

	return eventErrors, nil
}

func (s *service) GetTransfers(ctx context.Context, transferIDs []string) ([]Transfer, error) {
	tbTransferIDs := make([]tigerbeetleTypes.Uint128, len(transferIDs))
	for i, id := range transferIDs {
		_, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("Transfer ID must be a uuid. %w", ErrInvalidArg)
		}
		transferID, err := UuidToU128(id)
		if err != nil {
			return nil, err
		}

		tbTransferIDs[i] = *transferID
	}

	results, err := s.tb.LookupTransfers(tbTransferIDs)
	if err != nil {
		return nil, err
	}

	ret := make([]Transfer, len(results))
	for i, transfer := range results {
		ret[i] = Transfer{
			ID:              U128ToUuid(transfer.ID),
			DebitAccountID:  U128ToUuid(transfer.DebitAccountID),
			CreditAccountID: U128ToUuid(transfer.CreditAccountID),
			Amount:          transfer.Amount,
			Code:            transfer.Code,
			Flags: TransferFlags{
				Linked:         transfer.Flags&(1<<0) == 1,
				TwoPhaseCommit: transfer.Flags&(1<<1) == 2,
				Condition:      transfer.Flags&(1<<2) == 4,
			},
		}
	}

	// TODO: ACL on ledgers involved

	return ret, nil
}

type CommitTransferArgs struct {
	ID    string `validate:"required,uuid4"`
	Flags CommitFlags
	Code  uint32
}

func (s *service) CommitTransfers(ctx context.Context, args []CommitTransferArgs) ([]EventResult, error) {
	commits := make([]tigerbeetleTypes.Commit, len(args))
	for i, commit := range args {
		err := s.validator.Struct(commit)
		if err != nil {
			return nil, fmt.Errorf("%s %w", err.Error(), ErrInvalidArg)
		}

		commitID, err := UuidToU128(commit.ID)
		if err != nil {
			return nil, fmt.Errorf("%s %w", err.Error(), ErrInternal)
		}

		commits[i] = tigerbeetleTypes.Commit{
			ID:    *commitID,
			Code:  commit.Code,
			Flags: commit.Flags.ToUint32(),
		}
	}

	eventErrors, err := s.tb.CommitTransfers(commits)
	if err != nil {
		return nil, err
	}

	return eventErrors, nil
}
