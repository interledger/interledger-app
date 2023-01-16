package client_test

import (
	"context"
	"log"
	"os"
	"testing"

	"gitlab.com/fynbos/backend/statements"
	"gitlab.com/fynbos/backend/statements/client"
)

func TestGenerateWalletStatement(t *testing.T) {
	client := client.New()
	pdf, err := client.GenerateWalletStatementPDF(context.Background(), statements.GenerateWalletStatementArgs{
		Name:        "Alice Smith",
		AccountID:   "19964",
		Period:      "2022-01-01",
		BalanceDate: "Jan 31 2022",
		Balance:     "$100.00",
		Transactions: []statements.TransactionTableRow{
			{
				Date:        "Jan 10 2022",
				Description: "Payment to cash balance",
				Amount:      "$10.00",
				RecieptID:   "4051-8073",
			},
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
