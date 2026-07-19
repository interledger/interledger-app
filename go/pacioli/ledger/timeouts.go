package ledger

import (
	"context"
	"math/rand"
	"time"

	"github.com/interledger/interledger-app/go/log"
	"go.uber.org/zap"

	"github.com/interledger/interledger-app/go/pacioli/ledger/tigerroach"
)

func TimeoutTransfersForever(b Backends) {
	rand.New(rand.NewSource(time.Now().UnixNano()))
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

	return nil
}
