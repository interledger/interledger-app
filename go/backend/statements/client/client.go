package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/signintech/gopdf"
	"gitlab.com/fynbos/backend/providers/machnet"
	"gitlab.com/fynbos/backend/statements"
	"gitlab.com/fynbos/backend/transactions"
)

var _ statements.Client = client{}

var (
	fontDisplay    = "font-display"
	fontDisplayUrl = "https://cdn.fynbos.app/fonts/poppins/v20/400/Regular.ttf"

	fontInter    = "font-inter"
	fontInterUrl = "https://cdn.fynbos.app/fonts/inter/v12/Regular.ttf"

	xPixelWidth  = 1190.00    // see figma
	yPixelHeight = 1684.00    // see figma
	xLeftMargin  = xunit(124) // see figma
	yTopMargin   = yunit(150) // see figma

	textColourMedium = rgbColour{R: 51, G: 65, B: 85}
	textColourStrong = rgbColour{R: 15, G: 23, B: 42}

	tableEvenRowBackground = rgbColour{R: 248, G: 250, B: 252}
	tableOddRowBackground  = rgbColour{R: 255, G: 255, B: 255}
)

func xunit(value float64) float64 {
	return value * gopdf.PageSizeA4.W / xPixelWidth
}

func yunit(value float64) float64 {
	return value * gopdf.PageSizeA4.H / yPixelHeight
}

type client struct{}

func New() *client {
	return &client{}
}

func (c client) GenerateWalletStatementPDF(ctx context.Context, wallet *machnet.Wallet, transactions []transactions.Transaction) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.SetMargins(xLeftMargin, yTopMargin, gopdf.PageSizeA4.W-xLeftMargin, gopdf.PageSize10x14.H-yTopMargin)
	err := loadFont(&pdf, fontDisplay, fontDisplayUrl)
	if err != nil {
		return nil, err
	}
	err = loadFont(&pdf, fontInter, fontInterUrl)
	if err != nil {
		return nil, err
	}

	startNewPage(&pdf)
	pdf.SetX(pdf.MarginLeft())
	err = pdf.SetFont(fontInter, "", 14)
	if err != nil {
		return nil, err
	}
	err = tableFromRows(&pdf, tableFromRowsArgs{
		RowHeight: yunit(40),
		ColWidths: []float64{xunit(160), xunit(260), xunit(380), xunit(144)},
		Headers: tableRow{
			FontSize:   float64(8),
			TextColour: textColourMedium,
			Entries:    []string{"Date", "Receipt number", "Description", "Date"},
		},
		Rows: []tableRow{
			{
				FontSize:   float64(8),
				TextColour: textColourStrong,
				Entries:    []string{"Jan 10 2022", "4051-8073", "Payment to cash balance", "$10.00"},
			},
			{
				FontSize:   float64(8),
				TextColour: textColourStrong,
				Entries:    []string{"Jan 10 2022", "4051-8073", "Payment to cash balance", "$10.00"},
				TextAlign:  gopdf.Middle,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return pdf.GetBytesPdfReturnErr()
}

type rgbColour struct {
	R uint8
	G uint8
	B uint8
}

type cellInfo struct {
	FontSize   float64
	Align      int
	TextColour rgbColour
	Text       string
}

type tableFromRowsArgs struct {
	RowHeight float64
	ColWidths []float64
	Rows      []tableRow
	Headers   tableRow
}

type tableRow struct {
	Entries          []string
	TextColour       rgbColour
	TextAlign        int
	FontSize         float64
	LinkedToPrevious bool
}

func tableFromRows(pdf *gopdf.GoPdf, args tableFromRowsArgs) error {
	// check table structure
	for i, row := range args.Rows {
		if len(row.Entries) != len(args.ColWidths) {
			return errors.New(fmt.Sprintf("pdf error: row %d does not have %d columns.", i, len(args.ColWidths)))
		}
	}

	err := addRow(pdf, args.RowHeight, args.ColWidths, tableOddRowBackground, args.Headers)
	if err != nil {
		return err
	}

	pdf.SetX(pdf.MarginLeft())
	err = addRow(pdf, args.RowHeight, args.ColWidths, tableEvenRowBackground, args.Rows[0])
	if err != nil {
		return err
	}

	pdf.SetX(pdf.MarginLeft())
	err = addRow(pdf, args.RowHeight, args.ColWidths, tableOddRowBackground, args.Rows[1])
	if err != nil {
		return err
	}

	return nil
}

func addRow(pdf *gopdf.GoPdf, rowHeight float64, colWidths []float64, fill rgbColour, row tableRow) error {
	err := pdf.SetFontSize(row.FontSize)
	if err != nil {
		return err
	}
	pdf.SetTextColor(row.TextColour.R, row.TextColour.G, row.TextColour.B)
	for i, entry := range row.Entries {
		carriagePosition := gopdf.Right
		if i == (len(row.Entries) - 1) {
			carriagePosition = gopdf.Bottom
		}

		err = pdf.CellWithOption(&gopdf.Rect{
			H: rowHeight,
			W: colWidths[i],
		}, entry, gopdf.CellOption{
			Align: row.TextAlign,
			Float: carriagePosition,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func startNewPage(pdf *gopdf.GoPdf) {
	pdf.AddPage()
	template := pdf.ImportPage("walletStatementA4.pdf", pdf.GetNumberOfPages(), "/MediaBox")
	pdf.UseImportedTemplate(template, 0, 0, 0, 0)
}

func loadFont(pdf *gopdf.GoPdf, name, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	err = pdf.AddTTFFontByReader(name, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
