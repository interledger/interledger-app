package client_test

import (
	"context"
	"os"
	"testing"

	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/statements/client"
	"gitlab.com/fynbos/backend/transactions"
)

func TestGenerateWalletStatement(t *testing.T) {
	t.Skip("Needs wkhtmltopdf to be installed on your machine.")
	statements := client.New()

	pdf, err := statements.GenerateWalletStatementPDF(context.Background(), &machnet.Wallet{
		AvailableBalance: 1000,
	}, []transactions.Transaction{})
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile("./statement.pdf", pdf, 0644)
	if err != nil {
		t.Fatal(err)
	}
}
