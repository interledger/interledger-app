package jobs

import (
	"context"
	"errors"
	"time"

	"go.temporal.io/sdk/activity"

	"github.com/interledger/interledger-app/go/backend/wallets"

	"go.temporal.io/sdk/workflow"
)

func MigratePaymentPointers(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("MigratePaymentPointers workflow started")

	return workflow.ExecuteActivity(ctx, a.MigratePaymentPointers).Get(ctx, nil)
}

func (a *Activity) MigratePaymentPointers(ctx context.Context) error {
	type PP struct {
		WalletID string `db:"wallet_id"`
		URL      string `db:"url"`
	}

	var AllPP []PP
	err := a.b.DB().SelectContext(ctx, &AllPP, "SELECT wallet_id, url FROM payment_pointers")
	if err != nil {
		return err
	}

	var createdCnt, skipCnt int
	for _, pp := range AllPP {
		_, err = a.b.Wallets().AddAddress(ctx, pp.WalletID, pp.URL)
		if errors.Is(err, wallets.ErrDuplicateWallet) {
			skipCnt++
			continue
		}
		if err != nil {
			return err
		}
		createdCnt++
	}

	logger := activity.GetLogger(ctx)
	logger.Info("Payment Pointer Migration", "payment_pointer_cnt", len(AllPP), "wallets_created", createdCnt, "wallets_already_exits", skipCnt)

	return nil
}

func MigrateOpenPaymentsObjects(ctx workflow.Context) error {
	var a *Activity
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("MigratePaymentPointers workflow started")

	err := workflow.ExecuteActivity(ctx, a.MigrateOpenPaymentsQuotes).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.MigrateOpenPaymentsIncomingPayments).Get(ctx, nil)
	if err != nil {
		return err
	}

	err = workflow.ExecuteActivity(ctx, a.MigrateOpenPaymentsOutgoingPayments).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func (a *Activity) MigrateOpenPaymentsOutgoingPayments(ctx context.Context) error {
	type PP struct {
		ID  string `db:"id"`
		URL string `db:"url"`
	}

	type OutgoingPayment struct {
		ID       string `db:"id"`
		SenderPP string `db:"send_payment_pointer_id"`
		RecvPP   string `db:"to_payment_pointer_id"`
	}

	var AllPP []PP
	err := a.b.DB().SelectContext(ctx, &AllPP, "SELECT id, url FROM payment_pointers")
	if err != nil {
		return err
	}

	ppMap := make(map[string]string)
	for _, pp := range AllPP {
		ppMap[pp.ID] = pp.URL
	}

	var AllOutgoing []OutgoingPayment
	err = a.b.DB().SelectContext(ctx, &AllOutgoing, "SELECT op.id, op.to_payment_pointer_id, q.send_payment_pointer_id FROM openpayments_outgoing_payment op INNER JOIN openpayments_quotes q ON q.id = op.quote_id")
	if err != nil {
		return err
	}

	uq, err := a.b.DB().PreparexContext(ctx, "UPDATE openpayments_outgoing_payment SET sender_wallet_address=$1, receiver_wallet_address=$2 WHERE id=$3")
	if err != nil {
		return err
	}
	defer uq.Close()

	var updateCnt int
	for _, op := range AllOutgoing {
		sendPP, ok := ppMap[op.SenderPP]
		if !ok {
			return errors.New("failed to lookup sender Payment Pointer")
		}

		recvPP, ok := ppMap[op.RecvPP]
		if !ok {
			return errors.New("failed to lookup receiver Payment Pointer")
		}

		_, err = uq.ExecContext(ctx, sendPP, recvPP, op.ID)
		if err != nil {
			return err
		}

		updateCnt++
	}

	logger := activity.GetLogger(ctx)
	logger.Info("Open payments outgoing payments migration", "updateCnt", updateCnt)

	return nil
}

func (a *Activity) MigrateOpenPaymentsIncomingPayments(ctx context.Context) error {
	type PP struct {
		ID  string `db:"id"`
		URL string `db:"url"`
	}

	type IncomingPayment struct {
		ID       string `db:"id"`
		SenderPP string `db:"from_payment_pointer_id"`
		RecvPP   string `db:"payment_pointer_id"`
	}

	var AllPP []PP
	err := a.b.DB().SelectContext(ctx, &AllPP, "SELECT id, url FROM payment_pointers")
	if err != nil {
		return err
	}

	ppMap := make(map[string]string)
	for _, pp := range AllPP {
		ppMap[pp.ID] = pp.URL
	}

	var AllIncoming []IncomingPayment
	err = a.b.DB().SelectContext(ctx, &AllIncoming, "SELECT id, from_payment_pointer_id, payment_pointer_id  FROM openpayments_incoming_payment")
	if err != nil {
		return err
	}

	uq, err := a.b.DB().PreparexContext(ctx, "UPDATE openpayments_incoming_payment SET sender_wallet_address=$1, receiver_wallet_address=$2 WHERE id=$3")
	if err != nil {
		return err
	}
	defer uq.Close()

	var updateCnt int
	for _, ip := range AllIncoming {
		sendPP, ok := ppMap[ip.SenderPP]
		if !ok {
			return errors.New("failed to lookup sender Payment Pointer")
		}

		recvPP, ok := ppMap[ip.RecvPP]
		if !ok {
			return errors.New("failed to lookup receiver Payment Pointer")
		}

		_, err = uq.ExecContext(ctx, sendPP, recvPP, ip.ID)
		if err != nil {
			return err
		}

		updateCnt++
	}

	logger := activity.GetLogger(ctx)
	logger.Info("Open payments incoming payments migration", "updateCnt", updateCnt)

	return nil
}

func (a *Activity) MigrateOpenPaymentsQuotes(ctx context.Context) error {
	type PP struct {
		ID  string `db:"id"`
		URL string `db:"url"`
	}

	type Quote struct {
		ID       string `db:"id"`
		SenderPP string `db:"send_payment_pointer_id"`
		RecvPP   string `db:"recv_payment_pointer_id"`
	}

	var AllPP []PP
	err := a.b.DB().SelectContext(ctx, &AllPP, "SELECT id, url FROM payment_pointers")
	if err != nil {
		return err
	}

	ppMap := make(map[string]string)
	for _, pp := range AllPP {
		ppMap[pp.ID] = pp.URL
	}

	var AllQuotes []Quote
	err = a.b.DB().SelectContext(ctx, &AllQuotes, "SELECT id, send_payment_pointer_id, recv_payment_pointer_id  FROM openpayments_quotes")
	if err != nil {
		return err
	}

	uq, err := a.b.DB().PreparexContext(ctx, "UPDATE openpayments_quotes SET sender_wallet_address=$1, receiver_wallet_address=$2 WHERE id=$3")
	if err != nil {
		return err
	}
	defer uq.Close()

	var updateCnt int
	for _, q := range AllQuotes {
		sendPP, ok := ppMap[q.SenderPP]
		if !ok {
			return errors.New("failed to lookup sender Payment Pointer")
		}

		recvPP, ok := ppMap[q.RecvPP]
		if !ok {
			return errors.New("failed to lookup receiver Payment Pointer")
		}

		_, err = uq.ExecContext(ctx, sendPP, recvPP, q.ID)
		if err != nil {
			return err
		}

		updateCnt++
	}

	logger := activity.GetLogger(ctx)
	logger.Info("Open payments quote migration", "updateCnt", updateCnt)

	return nil
}
