package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"

	"github.com/lightninglabs/lndclient"
)

func listenInvoicesForever(ctx context.Context, lc lnrpc.LightningClient, name string) {
	stream, err := lc.SubscribeInvoices(ctx, &lnrpc.InvoiceSubscription{
		AddIndex:    0,
		SettleIndex: 0,
	})
	if err != nil {
		panic(err)
	}

	for {
		fmt.Println(name, "listening for invoices")
		inv, err := stream.Recv()
		if err != nil {
			fmt.Println(name, err)
			if errors.Is(err, io.EOF) {
				return
			}
			panic(err)
		}

		fmt.Println(name, inv.Memo, inv.State, inv.Value)
	}
}

func listenTXForever(ctx context.Context, lc lnrpc.LightningClient, name string) {
	stream, err := lc.SubscribeTransactions(ctx, &lnrpc.GetTransactionsRequest{
		StartHeight: 0,
		EndHeight:   0,
	})
	if err != nil {
		panic(err)
	}

	for {
		fmt.Println(name, "listening for tx")
		tx, err := stream.Recv()
		if err != nil {
			fmt.Println(name, err)
			if errors.Is(err, io.EOF) {
				return
			}
			panic(err)
		}

		fmt.Println(name, tx.Amount, tx.BlockHeight, tx.DestAddresses)
	}
}

func main() {
	ctx := context.Background()
	alice, err := lndclient.NewBasicClient("127.0.0.1:10001", "/home/barnard/.polar/networks/1/volumes/lnd/alice/tls.cert", "/home/barnard/.polar/networks/1/volumes/lnd/alice/data/chain/bitcoin/regtest/", "regtest")
	if err != nil {
		panic(err)
	}

	go listenInvoicesForever(ctx, alice, "alice")
	go listenTXForever(ctx, alice, "alice")

	dave, err := lndclient.NewBasicClient("127.0.0.1:10004", "/home/barnard/.polar/networks/1/volumes/lnd/dave/tls.cert", "/home/barnard/.polar/networks/1/volumes/lnd/dave/data/chain/bitcoin/regtest/", "regtest")
	if err != nil {
		panic(err)
	}

	go listenInvoicesForever(ctx, dave, "dave")
	go listenTXForever(ctx, dave, "dave")
	/*
		txs, err := alice.GetTransactions(ctx, &lnrpc.GetTransactionsRequest{
			StartHeight: 0,
			EndHeight:   11000,
		})
		if err != nil {
			panic(err)
		}

		for _, tx := range txs.Transactions {
			fmt.Println(tx.Amount, tx.BlockHeight, tx.DestAddresses)
		}

		txs, err = dave.GetTransactions(ctx, &lnrpc.GetTransactionsRequest{
			StartHeight: 0,
			EndHeight:   1000,
		})
		if err != nil {
			panic(err)
		}

		for _, tx := range txs.Transactions {
			fmt.Println(tx.Amount, tx.BlockHeight, tx.DestAddresses)
		}
	*/
	chans, err := alice.ListChannels(ctx, &lnrpc.ListChannelsRequest{ActiveOnly: true})
	if err != nil {
		panic(err)
	}

	for _, ch := range chans.Channels {
		fmt.Println(ch.Capacity, ch.Initiator, ch.LocalBalance, ch.UnsettledBalance)
	}

	chans, err = dave.ListChannels(ctx, &lnrpc.ListChannelsRequest{ActiveOnly: true})
	if err != nil {
		panic(err)
	}

	for _, ch := range chans.Channels {
		fmt.Println(ch.Capacity, ch.Initiator, ch.LocalBalance, ch.UnsettledBalance)
	}

	inv, err := dave.AddInvoice(ctx, &lnrpc.Invoice{Memo: "What are we", Value: 5000})
	if err != nil {
		panic(err)
	}

	pay, err := alice.SendPaymentSync(ctx, &lnrpc.SendRequest{PaymentRequest: inv.PaymentRequest})
	if err != nil {
		panic(err)
	}

	fmt.Println(pay.PaymentError, pay.PaymentPreimage, pay.PaymentHash)

	chans, err = alice.ListChannels(ctx, &lnrpc.ListChannelsRequest{ActiveOnly: true})
	if err != nil {
		panic(err)
	}

	for _, ch := range chans.Channels {
		fmt.Println(ch.Capacity, ch.Initiator, ch.LocalBalance, ch.UnsettledBalance)
	}

	chans, err = dave.ListChannels(ctx, &lnrpc.ListChannelsRequest{ActiveOnly: true})
	if err != nil {
		panic(err)
	}

	for _, ch := range chans.Channels {
		fmt.Println(ch.Capacity, ch.Initiator, ch.LocalBalance, ch.UnsettledBalance)
	}

	time.Sleep(time.Minute * 10)
	/*
		inv, err := alice.AddInvoice(ctx, &lnrpc.Invoice{
			Memo:  "Hi Hi",
			Value: 5000,
		})

		fmt.Println(inv.PaymentRequest)

		alice.SendPaymentSync()
		chans, err = alice.ListChannels(ctx, &lnrpc.ListChannelsRequest{ActiveOnly: true})
		if err != nil {
			panic(err)
		}

		for _, ch := range chans.Channels {
			fmt.Println(ch.Capacity, ch.Initiator, ch.LocalBalance, ch.UnsettledBalance)
		}
	*/
}
