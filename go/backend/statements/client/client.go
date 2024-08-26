package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/signintech/gopdf"
	"gitlab.com/fynbos/backend/statements"
)

var _ statements.Client = client{}

var (
	walletStatementUrl = "https://cdn.fynbos.app/pdfs/statement-plain-A4.pdf"

	fontDisplay    = "font-display"
	fontDisplayUrl = "https://cdn.fynbos.app/fonts/poppins/v20/400/Regular.ttf"

	fontDisplayMedium    = "font-display-medium"
	fontDisplayMediumUrl = "https://cdn.fynbos.app/fonts/poppins/v20/500/Medium.ttf"

	fontInter    = "font-inter"
	fontInterUrl = "https://cdn.fynbos.app/fonts/inter/v12/Medium.ttf"

	fontInterMedium    = "font-inter"
	fontInterMediumUrl = "https://cdn.fynbos.app/fonts/inter/v12/Regular.ttf"

	xPixelWidth   = 1190.00    // see figma
	yPixelHeight  = 1684.00    // see figma
	xLeftMargin   = xunit(140) // see figma
	xRightMargin  = gopdf.PageSizeA4.W - xLeftMargin
	yTopMargin    = yunit(150)                     // see figma
	yBottomMargin = gopdf.PageSizeA4.H - yunit(60) // see figma

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

func pxToPt(value float64) float64 {
	return value * 0.75
}

type client struct{}

func New() *client {
	return &client{}
}

func (c client) GenerateWalletStatementPDF(ctx context.Context, args statements.GenerateWalletStatementArgs) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.SetMargins(xLeftMargin, yTopMargin, xRightMargin, yBottomMargin)
	err := loadFont(&pdf, fontDisplay, fontDisplayUrl)
	if err != nil {
		return nil, err
	}
	err = loadFont(&pdf, fontDisplayMedium, fontDisplayMediumUrl)
	if err != nil {
		return nil, err
	}
	err = loadFont(&pdf, fontInter, fontInterUrl)
	if err != nil {
		return nil, err
	}
	err = loadFont(&pdf, fontInterMedium, fontInterMediumUrl)
	if err != nil {
		return nil, err
	}

	err = startNewPage(&pdf)
	if err != nil {
		return nil, err
	}

	err = addWalletStatementHeader(&pdf, statementHeader{
		Name:          args.Name,
		AccountID:     args.AccountID,
		Balance:       args.Balance,
		Period:        args.Period,
		StatementDate: args.BalanceDate,
	})
	if err != nil {
		return nil, err
	}

	pdf.Br(20)

	err = pdf.SetFont(fontInter, "", 14)
	if err != nil {
		return nil, err
	}
	pdf.SetX(pdf.MarginLeft())
	rows := make([]tableRow, len(args.Transactions))
	var prevReceiptID string
	rowBackgroundColours := []rgbColour{
		tableOddRowBackground,
		tableEvenRowBackground,
	}
	rowColour := 0
	textAlignment := []int{gopdf.Middle, gopdf.Middle, gopdf.Middle, gopdf.Middle | gopdf.Right} // right align amount
	for i, trx := range args.Transactions {
		entries := []string{trx.Date, trx.RecieptID, trx.Description, trx.Amount}
		if prevReceiptID == trx.RecieptID {
			entries[0] = ""
			entries[1] = ""
		} else {
			rowColour++
		}
		prevReceiptID = trx.RecieptID
		rows[i] = tableRow{
			FontSize:         pxToPt(8),
			TextColour:       textColourStrong,
			TextAlign:        textAlignment,
			Entries:          entries,
			BackgroundColour: rowBackgroundColours[rowColour%2],
		}
	}
	err = tableFromRows(&pdf, tableFromRowsArgs{
		RowHeight: yunit(40),
		ColWidths: []float64{xunit(120), xunit(300), xunit(320), xunit(100)},
		Headers: tableRow{
			FontSize:   pxToPt(8),
			TextColour: textColourMedium,
			TextAlign:  textAlignment,
			Entries:    []string{"Date", "Receipt number", "Description", "Amount  "},
		},
		Rows: rows,
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

type tableFromRowsArgs struct {
	RowHeight float64
	ColWidths []float64
	Rows      []tableRow
	Headers   tableRow
}

type tableRow struct {
	Entries          []string
	TextColour       rgbColour
	TextAlign        []int
	FontSize         float64
	LinkedToPrevious bool
	BackgroundColour rgbColour
}

func tableFromRows(pdf *gopdf.GoPdf, args tableFromRowsArgs) error {
	// check table structure
	for i, row := range args.Rows {
		if len(row.Entries) != len(args.ColWidths) {
			return fmt.Errorf("pdf error: row %d does not have %d columns.", i, len(args.ColWidths))
		}
	}

	err := addRow(pdf, args.RowHeight, args.ColWidths, tableOddRowBackground, args.Headers)
	if err != nil {
		return err
	}

	for _, row := range args.Rows {
		if pdf.GetY()+args.RowHeight > pdf.MarginBottom() {
			err = startNewPage(pdf)
			if err != nil {
				return err
			}

			err := addRow(pdf, args.RowHeight, args.ColWidths, tableOddRowBackground, args.Headers)
			if err != nil {
				return err
			}
		}

		err = addRow(pdf, args.RowHeight, args.ColWidths, row.BackgroundColour, row)
		if err != nil {
			return err
		}
	}

	return nil
}

func addRow(pdf *gopdf.GoPdf, rowHeight float64, colWidths []float64, fill rgbColour, row tableRow) error {
	err := pdf.SetFontSize(row.FontSize)
	if err != nil {
		return err
	}

	// row background
	pdf.SetX(pdf.MarginLeft())
	pdf.SetTextColor(row.TextColour.R, row.TextColour.G, row.TextColour.B)
	pdf.SetFillColor(fill.R, fill.G, fill.B)
	pdf.RectFromUpperLeftWithStyle(pdf.GetX(), pdf.GetY(), xRightMargin-xLeftMargin, rowHeight, "F")

	pdf.SetX(pdf.MarginLeft() + pxToPt(4))
	for i, entry := range row.Entries {
		carriagePosition := gopdf.Right
		if i == (len(row.Entries) - 1) {
			carriagePosition = gopdf.Bottom
		}

		err = pdf.CellWithOption(&gopdf.Rect{
			H: rowHeight,
			W: colWidths[i],
		}, entry, gopdf.CellOption{
			Align: row.TextAlign[i],
			Float: carriagePosition,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func startNewPage(pdf *gopdf.GoPdf) error {
	pdf.AddPage()
	resp, err := http.Get(walletStatementUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// hack for now. Can't get ImportPageStream to work
	file, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		return err
	}

	_, err = file.Write(body)
	if err != nil {
		return err
	}

	template := pdf.ImportPage(file.Name(), 1, "/MediaBox")
	pdf.UseImportedTemplate(template, 0, 0, 0, 0)

	// write page number
	pdf.SetXY(pdf.MarginLeft(), pdf.MarginBottom())
	err = pdf.SetFont(fontInter, "", pxToPt(8))
	if err != nil {
		return err
	}
	err = pdf.CellWithOption(&gopdf.Rect{
		H: yunit(40),
		W: pdf.MarginRight() - pdf.MarginLeft(),
	}, fmt.Sprintf("%d", pdf.GetNumberOfPages()), gopdf.CellOption{
		Align: gopdf.Middle | gopdf.Center,
		Float: gopdf.Right,
	})
	if err != nil {
		return err
	}

	// reset to top left
	pdf.SetXY(pdf.MarginLeft(), pdf.MarginTop())

	return nil
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

type statementHeader struct {
	Name          string
	AccountID     string
	Balance       string
	Period        string
	StatementDate string
}

func addWalletStatementHeader(pdf *gopdf.GoPdf, header statementHeader) error {
	pdf.SetTextColor(textColourStrong.R, textColourStrong.G, textColourStrong.B)
	err := pdf.SetFont(fontDisplayMedium, "", pxToPt(18))
	if err != nil {
		return err
	}

	cellWidth := (gopdf.PageSizeA4.W - xLeftMargin) / 3
	cellHeight := yunit(40)
	err = pdf.CellWithOption(&gopdf.Rect{
		H: cellHeight,
		W: cellWidth,
	}, header.Name, gopdf.CellOption{
		Align: gopdf.Middle,
		Float: gopdf.Right,
	})
	if err != nil {
		return err
	}

	err = pdf.CellWithOption(&gopdf.Rect{
		H: cellHeight,
		W: cellWidth,
	}, "", gopdf.CellOption{
		Align: gopdf.Bottom,
		Float: gopdf.Right,
	})
	if err != nil {
		return err
	}

	pdf.SetTextColor(textColourMedium.R, textColourMedium.G, textColourMedium.B)
	err = pdf.SetFont(fontInter, "", pxToPt(8))
	if err != nil {
		return err
	}
	err = pdf.CellWithOption(&gopdf.Rect{
		H: cellHeight,
		W: cellWidth,
	}, fmt.Sprintf("Cash balance as at %s", header.StatementDate), gopdf.CellOption{
		Align: gopdf.Bottom,
		Float: gopdf.Bottom,
	})
	if err != nil {
		return err
	}

	pdf.SetX(pdf.MarginLeft())
	pdf.SetTextColor(textColourStrong.R, textColourStrong.G, textColourStrong.B)
	err = pdf.CellWithOption(&gopdf.Rect{
		H: cellHeight,
		W: cellWidth,
	}, fmt.Sprintf("AccountID: %s", header.AccountID), gopdf.CellOption{
		Align: gopdf.Top,
		Float: gopdf.Right,
	})
	if err != nil {
		return err
	}

	err = pdf.CellWithOption(&gopdf.Rect{
		H: cellHeight,
		W: cellWidth,
	}, "", gopdf.CellOption{
		Align: gopdf.Bottom,
		Float: gopdf.Right,
	})
	if err != nil {
		return err
	}

	pdf.SetTextColor(textColourStrong.R, textColourStrong.G, textColourStrong.B)
	err = pdf.SetFont(fontDisplayMedium, "", pxToPt(18))
	if err != nil {
		return err
	}
	err = pdf.CellWithOption(&gopdf.Rect{
		H: cellHeight,
		W: cellWidth,
	}, header.Balance, gopdf.CellOption{
		Align: gopdf.Middle,
		Float: gopdf.Bottom,
	})
	if err != nil {
		return err
	}

	pdf.Br(10)
	pdf.SetX(pdf.MarginLeft())
	pdf.SetTextColor(textColourMedium.R, textColourMedium.G, textColourMedium.B)
	err = pdf.SetFont(fontInter, "", pxToPt(8))
	if err != nil {
		return err
	}
	err = pdf.CellWithOption(&gopdf.Rect{
		H: cellHeight,
		W: cellWidth,
	}, "Statement summary", gopdf.CellOption{
		Align: gopdf.Bottom,
		Float: gopdf.Bottom,
	})
	if err != nil {
		return err
	}

	pdf.SetX(pdf.MarginLeft())
	pdf.SetTextColor(textColourStrong.R, textColourStrong.G, textColourStrong.B)
	err = pdf.SetFont(fontInterMedium, "", pxToPt(8))
	if err != nil {
		return err
	}
	err = pdf.CellWithOption(&gopdf.Rect{
		H: cellHeight,
		W: cellWidth,
	}, header.Period, gopdf.CellOption{
		Align: gopdf.Middle,
		Float: gopdf.Bottom,
	})
	if err != nil {
		return err
	}

	return nil
}
