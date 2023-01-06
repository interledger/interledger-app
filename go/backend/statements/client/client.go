package client

import (
	"bytes"
	"context"
	"embed"
	"html/template"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/statements"
	"gitlab.com/fynbos/backend/transactions"
)

var _ statements.Client = client{}

type client struct{}

func New() *client {
	return &client{}
}

//go:embed templates/*.html
var templates embed.FS

func (c client) GenerateWalletStatementPDF(ctx context.Context, wallet *machnet.Wallet, transactions []transactions.Transaction) ([]byte, error) {
	tmpl, err := template.ParseFS(templates, "templates/walletstatement.html")
	if err != nil {
		return nil, err
	}

	data := struct {
		AvailableBalance uint64
	}{
		AvailableBalance: wallet.AvailableBalance,
	}
	var html bytes.Buffer
	err = tmpl.Execute(&html, data)
	if err != nil {
		return nil, err
	}

	pdfg, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, err
	}

	page := wkhtmltopdf.NewPageReader(bytes.NewReader(html.Bytes()))

	// enable this if the HTML file contains local references such as images, CSS, etc.
	page.EnableLocalFileAccess.Set(true)

	// add the page to your generator
	pdfg.AddPage(page)

	// manipulate page attributes as needed
	pdfg.MarginLeft.Set(0)
	pdfg.MarginRight.Set(0)
	pdfg.Dpi.Set(300)
	pdfg.PageSize.Set(wkhtmltopdf.PageSizeA4)
	pdfg.Orientation.Set(wkhtmltopdf.OrientationLandscape)

	err = pdfg.Create()
	if err != nil {
		return nil, err
	}

	return pdfg.Bytes(), nil
}
