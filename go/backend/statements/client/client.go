package client

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/signintech/gopdf"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/statements"
	"gitlab.com/fynbos/backend/transactions"
)

var _ statements.Client = client{}

type client struct{}

func New() *client {
	return &client{}
}

func (c client) GenerateWalletStatementPDF(ctx context.Context, wallet *machnet.Wallet, transactions []transactions.Transaction) ([]byte, error) {

	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4}) //595.28, 841.89 = A4
	pdf.AddPage()

	// TODO: add hosted font
	err := pdf.AddTTFFont("Roboto", "./Roboto-Regular.ttf")
	if err != nil {
		return nil, err
	}

	err = pdf.SetFont("Roboto", "", 14)
	if err != nil {
		return nil, err
	}

	logo, err := getLogoImageHolder()
	if err != nil {
		return nil, err
	}

	err = pdf.ImageByHolder(logo, 0, 0, nil)
	if err != nil {
		return nil, err
	}
	pdf.Br(20)

	err = pdf.Cell(nil, fmt.Sprintf("Balance: %d", wallet.AvailableBalance))
	if err != nil {
		return nil, err
	}
	pdf.Br(20)

	for _, trx := range transactions {
		err = pdf.Cell(nil, fmt.Sprintf("%s: %s", trx.Timestamp.Format("2006-02-01"), trx.Amount.FormatAmount()))
		if err != nil {
			return nil, err
		}
		pdf.Br(20)
	}

	return pdf.GetBytesPdfReturnErr()
}

func getLogoImageHolder() (gopdf.ImageHolder, error) {
	resp, err := http.Get("https://cdn.fynbos.workers.dev/logos/32px.png")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	holder, err := gopdf.ImageHolderByBytes(imageBytes)
	if err != nil {
		return nil, err
	}

	return holder, nil
}
