package ledger

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/fynbos/pacioli"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	tigerbeetleTypes "github.com/coilhq/tigerbeetle-go/pkg/types"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

/*type Service interface {

	// This is declaritive and will not fail if the ledger exists. It will fail if one exists with
	// different fields.
	ConfigureLedgers(ctx context.Context, args []pacioli.ConfigureLedgerArgs) ([]pacioli.EventResult, error)
	GetLedgers(ctx context.Context, ledgerIDs []uint32) ([]pacioli.Ledger, error)

	// This is declaritive and will not fail if the account exists. It will fail if one exists with
	// different fields.
	ConfigureAccounts(ctx context.Context, args []pacioli.ConfigureAccountArgs) ([]pacioli.AccountResult, error)
	GetAccounts(ctx context.Context, accountIDs []string) ([]pacioli.Account, error)
	CreateTransfers(ctx context.Context, args []pacioli.CreateTransferArgs) ([]pacioli.TransferResult, error)
	GetTransfers(ctx context.Context, transferIDs []string) ([]pacioli.Transfer, error)
	CommitTransfers(ctx context.Context, transferIDs []string) ([]pacioli.TransferResult, error)
	VoidTransfers(ctx context.Context, transferIDs []string) ([]pacioli.TransferResult, error)
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
		return nil, fmt.Errorf("%s %w", err.Error(), pacioli.ErrInvalidArg)
	}

	return &service{db: args.Db, tb: args.Tb, validator: validator}, nil
}*/

// This is declaritive and will not fail if the ledger exists. It will fail if one exists with
// different fields.
func ConfigureLedgers(
	ctx context.Context,
	b Backends,
	args []pacioli.ConfigureLedgerArgs,
) ([]pacioli.EventResult, error) {
	ledgerIds := make([]uint32, len(args))
	for i, ledger := range args {
		err := b.Validator().Struct(ledger)
		if err != nil {
			return nil, fmt.Errorf("%s %d %s %w", "index: ", i, err.Error(), pacioli.ErrInvalidArg)
		}

		ledgerIds[i] = ledger.ID
	}

	existingLedgers, err := GetLedgers(ctx, b, ledgerIds)
	if err != nil {
		return nil, err
	}

	errorEvents := []pacioli.EventResult{}
	createdLedgers := []pacioli.Ledger{}
	err = crdbsqlx.ExecuteTx(ctx, b.DB(), nil, func(tx *sqlx.Tx) error {
		for i, ledger := range args {
			exists := false
			for _, existing := range existingLedgers {
				if ledger.ID == existing.ID {
					exists = true
					result := canCreateLedger(ledger, existing)
					if result != pacioli.LEDGER_OK {
						errorEvents = append(errorEvents, pacioli.EventResult{
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
					if result != pacioli.LEDGER_OK {
						errorEvents = append(errorEvents, pacioli.EventResult{
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
					return fmt.Errorf("%s %d %s %w", "index: ", i, err.Error(), pacioli.ErrInternal)
				}

				createdLedgers = append(createdLedgers, pacioli.Ledger{
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

func canCreateLedger(args pacioli.ConfigureLedgerArgs, existingLedger pacioli.Ledger) uint8 {
	if args.Name != existingLedger.Name {
		return pacioli.LEDGER_EXISTS_WITH_DIFFERENT_NAME
	}

	if args.Asset != existingLedger.Asset {
		return pacioli.LEDGER_EXISTS_WITH_DIFFERENT_ASSET
	}

	if args.Scale != existingLedger.Scale {
		return pacioli.LEDGER_EXISTS_WITH_DIFFERENT_SCALE
	}

	return pacioli.LEDGER_OK
}

func GetLedgers(ctx context.Context, b Backends, ids []uint32) ([]pacioli.Ledger, error) {
	// TODO: ACL

	var ledgers []pacioli.Ledger
	query, args, err := sqlx.In("SELECT * FROM ledgers WHERE id IN (?);", ids)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), pacioli.ErrInternal)
	}
	err = b.DB().SelectContext(ctx, &ledgers, b.DB().Rebind(query), args...)
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), pacioli.ErrInternal)
	}

	return ledgers, nil
}

// Helper function to convert uuids into u128 needed for TigerBeetle IDs.
// TODO: see if there is a better way to do this.
func UuidToU128(value string) (*tigerbeetleTypes.Uint128, error) {
	src := strings.Replace(value, "-", "", -1)
	var ret tigerbeetleTypes.Uint128
	bytesWritten, err := hex.Decode(ret[:], []byte(src))
	if err != nil {
		return nil, err
	}
	if bytesWritten > 16 {
		return nil, errors.New("String could not be converted into uint128.")
	}

	return &ret, nil
}

// Helper function to extract the uuid we put into the u128.
func U128ToUuid(value tigerbeetleTypes.Uint128) string {
	s := hex.EncodeToString(value[:])
	ret := s[:8] + "-" + s[8:12] + "-" + s[12:16] + "-" + s[16:20] + "-" + s[20:]

	return ret
}

// This is declaritive and will not fail if the account exists. It will fail if one exists with
// different fields.
func ConfigureAccounts(
	ctx context.Context,
	b Backends,
	args []pacioli.ConfigureAccountArgs,
) ([]pacioli.AccountResult, error) {
	// TODO: ACL
	ledgerIDs := []uint32{}
	keys := map[uint32]uint8{}
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

	ledgers, err := GetLedgers(ctx, b, ledgerIDs)
	if err != nil {
		return nil, err
	}

	// mark these ledgers as existing.
	for _, ledger := range ledgers {
		keys[ledger.ID] = EXISTS
	}

	// size to len(args) to avoid appending
	eventErrors := make([]pacioli.AccountResult, len(args))
	errs := uint32(0) // number of errors
	accountsToCreate := make([]tigerbeetleTypes.Account, len(args))
	index := uint32(0) // number of accounts to create

	// stores the mapping from the tb create account args to the original args
	mapToEventErrorSlot := map[uint32]uint32{}
	for i, acc := range args {
		err := b.Validator().Struct(acc)
		if err != nil {
			return nil, fmt.Errorf("%s %d %s %w", "index: ", i, err.Error(), pacioli.ErrInvalidArg)
		}

		if keys[acc.LedgerID] != EXISTS {
			eventErrors[errs] = pacioli.AccountResult{
				Index: uint32(i),
				Code:  tigerbeetleTypes.CreateAccountResult(pacioli.ACCOUNT_LEDGER_DOES_NOT_EXIST),
			}
			errs++
			continue
		}

		tbAccID, err := UuidToU128(acc.ID)
		if err != nil {
			return nil, err
		}
		accountsToCreate[index] = tigerbeetleTypes.Account{
			ID:     *tbAccID,
			Code:   acc.Code,
			Flags:  acc.Flags.ToUint16(),
			Ledger: acc.LedgerID,
		}
		mapToEventErrorSlot[index] = uint32(i)
		index++
	}
	tbEventErrors, err := b.TigerBeetle().CreateAccounts(accountsToCreate[:index])
	// this error will be due to connection / io buffer issues
	if err != nil {
		return nil, fmt.Errorf("%s %w", err.Error(), pacioli.ErrInternal)
	}
	// map the tbEventErrors to our eventErrors
	for _, tbErr := range tbEventErrors {
		index, present := mapToEventErrorSlot[tbErr.Index]
		if !present {
			// the mapping is broken
			panic("Unable to map Tb event errs back to our create account errs.")
		}
		if tbErr.Code != tigerbeetleTypes.AccountExists {
			eventErrors[errs] = pacioli.AccountResult{Index: index, Code: tbErr.Code}
			errs++
		}
	}

	return eventErrors[:errs], nil
}

func GetAccounts(
	ctx context.Context,
	b Backends,
	accountIDs []string,
) ([]pacioli.Account, error) {
	tbAccIDs := []tigerbeetleTypes.Uint128{}
	for _, id := range accountIDs {
		_, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("Account ID must be a uuid. %w", pacioli.ErrInvalidArg)
		}
		accID, err := UuidToU128(id)
		if err != nil {
			return nil, err
		}

		tbAccIDs = append(tbAccIDs, *accID)
	}

	results, err := b.TigerBeetle().LookupAccounts(tbAccIDs)
	if err != nil {
		return nil, err
	}

	ret := make([]pacioli.Account, len(results))
	for i, result := range results {
		ret[i] = pacioli.Account{
			ID:             U128ToUuid(result.ID),
			LedgerID:       result.Ledger,
			Code:           result.Code,
			DebitsPending:  result.DebitsPending,
			DebitsPosted:   result.DebitsPosted,
			CreditsPending: result.CreditsPending,
			CreditsPosted:  result.CreditsPosted,
			Flags: pacioli.AccountFlags{
				Linked:                     result.Flags&(1<<0) == 1, // TODO: Create consts
				DebitsMustNotExceedCredits: result.Flags&(1<<1) == 2,
				CreditsMustNotExceedDebits: result.Flags&(1<<2) == 4,
			},
		}
	}

	// TODO: ACL on accounts

	return ret, nil
}

// TODO: Assuming that IDs are uuids and are being generated by another service for now. Might be better to be
// unopinionated and accept []byte.
func CreateTransfers(ctx context.Context, b Backends, args []pacioli.CreateTransferArgs) ([]pacioli.TransferResult, error) {
	// TODO: collect ledgers and perform ACL. TigerBeetle will introduce the ledger field onto
	// a transfer.
	tbTransfers := make([]tigerbeetleTypes.Transfer, len(args))
	for i, transfer := range args {
		err := b.Validator().Struct(transfer)
		if err != nil {
			return nil, fmt.Errorf("%s %w", err.Error(), pacioli.ErrInvalidArg)
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
			Flags:           transfer.Flags.ToUint16(),
			Timeout:         transfer.Timeout,
			Ledger:          transfer.Ledger,
		}
	}

	eventErrors, err := b.TigerBeetle().CreateTransfers(tbTransfers)
	if err != nil {
		return nil, err
	}

	return eventErrors, nil
}

func GetTransfers(ctx context.Context, b Backends, transferIDs []string) ([]pacioli.Transfer, error) {
	tbTransferIDs := make([]tigerbeetleTypes.Uint128, len(transferIDs))
	for i, id := range transferIDs {
		_, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("Transfer ID must be a uuid. %w", pacioli.ErrInvalidArg)
		}
		transferID, err := UuidToU128(id)
		if err != nil {
			return nil, err
		}

		tbTransferIDs[i] = *transferID
	}

	results, err := b.TigerBeetle().LookupTransfers(tbTransferIDs)
	if err != nil {
		return nil, err
	}

	ret := make([]pacioli.Transfer, len(results))
	for i, transfer := range results {
		ret[i] = pacioli.Transfer{
			ID:              U128ToUuid(transfer.ID),
			DebitAccountID:  U128ToUuid(transfer.DebitAccountID),
			CreditAccountID: U128ToUuid(transfer.CreditAccountID),
			Amount:          transfer.Amount,
			Code:            transfer.Code,
			Flags: pacioli.TransferFlags{
				Linked:              transfer.Flags&(1<<0) == 1,
				Pending:             transfer.Flags&(1<<1) == 2,
				PostPendingTransfer: transfer.Flags&(1<<2) == 4,
				VoidPendingTransfer: transfer.Flags&(1<<3) == 8,
			},
		}
	}

	// TODO: ACL on ledgers involved

	return ret, nil
}

func CommitTransfers(ctx context.Context, b Backends, transferIDs []string) ([]pacioli.TransferResult, error) {
	tbTransfers := make([]tigerbeetleTypes.Transfer, len(transferIDs))
	tbTransferIDs := make([]tigerbeetleTypes.Uint128, len(transferIDs))

	for i, id := range transferIDs {
		_, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("Transfer ID must be a uuid. %w", pacioli.ErrInvalidArg)
		}
		transferID, err := UuidToU128(id)
		if err != nil {
			return nil, err
		}

		tbTransferIDs[i] = *transferID
	}

	for i, tid := range tbTransferIDs {

		newID, err := UuidToU128(uuid.NewString())
		if err != nil {
			return nil, err
		}

		tbTransfers[i] = tigerbeetleTypes.Transfer{
			ID:        *newID,
			PendingID: tid,
			Flags: pacioli.TransferFlags{
				PostPendingTransfer: true,
			}.ToUint16(),
		}
	}

	eventErrors, err := b.TigerBeetle().CreateTransfers(tbTransfers)
	if err != nil {
		return nil, err
	}

	return eventErrors, nil
}

func VoidTransfers(ctx context.Context, b Backends, transferIDs []string) ([]pacioli.TransferResult, error) {
	tbTransfers := make([]tigerbeetleTypes.Transfer, len(transferIDs))
	tbTransferIDs := make([]tigerbeetleTypes.Uint128, len(transferIDs))

	for i, id := range transferIDs {
		_, err := uuid.Parse(id)
		if err != nil {
			return nil, fmt.Errorf("Transfer ID must be a uuid. %w", pacioli.ErrInvalidArg)
		}
		transferID, err := UuidToU128(id)
		if err != nil {
			return nil, err
		}

		tbTransferIDs[i] = *transferID
	}

	for i, tid := range tbTransferIDs {

		newID, err := UuidToU128(uuid.NewString())
		if err != nil {
			return nil, err
		}

		tbTransfers[i] = tigerbeetleTypes.Transfer{
			ID:        *newID,
			PendingID: tid,
			Flags: pacioli.TransferFlags{
				VoidPendingTransfer: true,
			}.ToUint16(),
		}
	}

	eventErrors, err := b.TigerBeetle().CreateTransfers(tbTransfers)
	if err != nil {
		return nil, err
	}

	return eventErrors, nil
}
