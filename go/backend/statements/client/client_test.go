package client_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/statements/client"
	"gitlab.com/fynbos/backend/transactions"
)

func TestGenerateWalletStatement(t *testing.T) {
	statements := client.New()

	pdf, err := statements.GenerateWalletStatementPDF(context.Background(), &machnet.Wallet{
		AvailableBalance: 1000,
	}, []transactions.Transaction{
		{
			Amount: currency.Amount{
				Value:    10,
				Currency: currency.Currency("USD"),
			},
			Timestamp: time.Now(),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	tempPdf, err := os.CreateTemp("", "tmpfile-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// comment out if you want to view the pdf.
		_ = os.Remove(tempPdf.Name())
	})

	if _, err := tempPdf.Write(pdf); err != nil {
		log.Fatal(err)
	}
}
