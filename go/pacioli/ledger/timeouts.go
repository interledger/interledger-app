package ledger

import (
	"context"
	"math/rand"
	"time"

	"gitlab.com/fynbos/log"
	"go.uber.org/zap"

	tb_types "github.com/coilhq/tigerbeetle-go/pkg/types"
	"github.com/google/uuid"

	"gitlab.com/fynbos/pacioli"
	"gitlab.com/fynbos/pacioli/ledger/tigerroach"
)

func TimeoutTransfersForever(b Backends) {
	rand.Seed(time.Now().UnixNano())
	standard := time.Minute * 5
	random := time.Duration(rand.Intn(220)) * time.Second
	ctx := context.Background()
	// Sleep a minimum of 5 minutes plus some random time.
	// This is so multiple processes don't try to timeout transfers at the same time, they will stagger.
	// Alternatively you'd need leader election, but this will do in the meantime.
	// Loop forever
	for {
		time.Sleep(standard + random)

		err := timeoutTransfers(ctx, b)
		if err != nil {
			log.Error("Error while trying to time out transfers", zap.Error(err))
		}
	}

}

func timeoutTransfers(ctx context.Context, b Backends) error {
	ctx, cancel := context.WithTimeout(ctx, time.Minute)
	defer cancel() // Don't leak

	ids, err := tigerroach.ListTimedoutTransferIDs(ctx, b)
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		// Nothing to do here.
		return nil
	}

	successIDs, err := tigerroach.TryTimeoutTransfers(ctx, b, ids)
	if err != nil {
		return err
	}

	if len(successIDs) == 0 {
		return nil
	}

	// Now attempt to do the same for tigerbeetle
	var tbTranfers []tb_types.Transfer
	for _, id := range successIDs {
		newID, err := UuidToU128(uuid.NewString())
		if err != nil {
			log.Error("failed to convert uuid for tigerbeetle", zap.Error(err))
			continue
		}

		pendingID, err := UuidToU128(id)
		if err != nil {
			log.Error("failed to convert uuid for tigerbeetle", zap.Error(err))
			continue
		}
		tbTranfers = append(tbTranfers, tb_types.Transfer{
			ID:        *newID,
			PendingID: *pendingID,
			Flags: pacioli.TransferFlags{
				VoidPendingTransfer: true,
			}.ToUint16(),
		})
	}
	tbErrors, err := b.TigerBeetle().CreateTransfers(tbTranfers)
	if err != nil {
		return err
	}

	var tbActualErrors []tb_types.TransferEventResult
	for _, tbErr := range tbErrors {
		switch tbErr.Code {
		case tb_types.TransferPendingTransferAlreadyVoided,
			tb_types.TransferPendingTransferExpired:
			continue
		}

		tbActualErrors = append(tbActualErrors, tbErr)
	}

	if len(tbActualErrors) > 0 {
		log.Warn("failed to sync timedout transfers to tigerbeetle with error codes",
			zap.Any("error_codes", tbActualErrors))
	}

	return nil
}
