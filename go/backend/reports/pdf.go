package reports

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
)

// PDF layout, in millimetres on landscape A4.
const (
	pdfMargin     = 10.0
	pdfFieldH     = 5.0
	pdfRowH       = 5.0
	pdfLabelW     = 46.0
	pdfValueW     = 92.5
	pdfFieldCols  = 2
	pdfFontFamily = "go"
	// pdfSectionMinH keeps a wallet's details from starting at the bottom of a page.
	pdfSectionMinH = 60.0
)

const pdfTableFontSize = 7.0

// pdfTxColumns are the transaction table's columns. Widths add up to the 277mm usable
// page width and were measured to fit the longest real values at pdfTableFontSize (a
// UUID, "web_monetization_outgoing", "payments_engine"); only title is truncated.
var pdfTxColumns = []struct {
	title string
	width float64
	align string
	value func(t reportTransaction) string
}{
	{"Transaction ID", 51, "L", func(t reportTransaction) string { return t.ID }},
	{"Created (UTC)", 27, "L", func(t reportTransaction) string { return pdfTime(t.CreatedAt) }},
	{"Updated (UTC)", 27, "L", func(t reportTransaction) string { return pdfTime(t.UpdatedAt) }},
	{"Type", 34, "L", func(t reportTransaction) string { return t.Type }},
	{"State", 16, "L", func(t reportTransaction) string { return t.State }},
	{"Provider", 23, "L", func(t reportTransaction) string { return t.Provider }},
	{"Amount", 20, "R", func(t reportTransaction) string { return t.Amount }},
	{"Fee", 14, "R", func(t reportTransaction) string { return t.Fee }},
	{"Currency", 13, "L", func(t reportTransaction) string { return t.Currency }},
	{"Title", 52, "L", func(t reportTransaction) string { return t.Title }},
}

func pdfTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02 15:04:05")
}

type pdfField struct {
	// column is the matching CSV column, which keeps the two formats in the same order.
	column string
	// label is the readable name shown in the PDF.
	label string
	value string
}

// fields returns the details shown above a wallet's transactions, in the same order
// as the CSV columns but with readable labels.
func (p providerReport) fields(u reportUser, w reportWallet) []pdfField {
	f := []pdfField{
		{"email", "Email", u.Email},
		{"first_name", "First name", u.FirstName},
		{"last_name", "Last name", u.LastName},
		{"user_id", "User ID", u.UserID},
		{"kratos_country", "Kratos country", u.KratosCountry},
		{"wallet_id", "Wallet ID", w.WalletID},
		{"wallet_country", "Wallet country", w.Country},
		{"local_kyc", "Local KYC *", w.LocalKYC},
	}
	if p.persona {
		f = append(f,
			pdfField{"persona_inquiry_id", "Persona inquiry ID", w.PersonaInquiryID},
			pdfField{"persona_inquiry_state", "Persona inquiry state", w.PersonaInquiryState},
		)
	}
	if p.externalUserID != nil {
		f = append(f, pdfField{p.name + "_user_id", p.displayName + " user ID", w.ExternalUserID})
	}
	return append(f,
		pdfField{p.name + "_linked_account_id", p.displayName + " linked account ID", w.LinkedAccountID},
		pdfField{"balance", "Balance", w.Balance},
		pdfField{"available_balance", "Available balance", w.AvailableBalance},
		pdfField{"balance_currency", "Balance currency", w.BalanceCurrency},
	)
}

// kycDisclaimer explains that Local KYC is not the KYC status of record.
func (p providerReport) kycDisclaimer() string {
	return "* Local KYC is this app's own record and may be out of date. The authoritative KYC status is held by " +
		p.displayName + "; check it there before relying on it."
}

// pdf renders the report with one section per wallet: the user's and wallet's details
// followed by the wallet's transaction history. Users without wallets get a section
// with the wallet fields left empty.
func (p providerReport) pdf(users []reportUser, now time.Time) ([]byte, error) {
	doc := p.newPDF(now)

	for _, u := range users {
		wallets := u.Wallets
		if len(wallets) == 0 {
			wallets = []reportWallet{{}}
		}
		for i, w := range wallets {
			p.pdfSection(doc, u, w, i == 0)
		}
	}

	if len(users) == 0 {
		doc.SetFont(pdfFontFamily, "", 10)
		doc.CellFormat(0, 10, "No users match this report.", "", 1, "L", false, 0, "")
	}

	var buf bytes.Buffer
	if err := doc.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (p providerReport) newPDF(now time.Time) *fpdf.Fpdf {
	doc := fpdf.New("L", "mm", "A4", "")
	doc.AddUTF8FontFromBytes(pdfFontFamily, "", goregular.TTF)
	doc.AddUTF8FontFromBytes(pdfFontFamily, "B", gobold.TTF)
	doc.SetTitle(p.title, true)
	doc.SetCreator("Interledger admin", true)
	doc.SetCreationDate(now)
	doc.SetMargins(pdfMargin, pdfMargin+8, pdfMargin)
	doc.SetAutoPageBreak(true, pdfMargin+6)
	doc.AliasNbPages("")

	generated := now.UTC().Format("2006-01-02 15:04 UTC")
	doc.SetHeaderFunc(func() {
		doc.SetY(pdfMargin)
		doc.SetFont(pdfFontFamily, "B", 11)
		doc.CellFormat(140, 6, p.title, "", 0, "L", false, 0, "")
		doc.SetFont(pdfFontFamily, "", 8)
		doc.SetTextColor(110, 110, 110)
		doc.CellFormat(0, 6, "Generated "+generated+" · Confidential: contains personal data", "", 1, "R", false, 0, "")
		doc.SetTextColor(0, 0, 0)
		doc.SetY(pdfMargin + 8)
	})
	doc.SetFooterFunc(func() {
		doc.SetY(-(pdfMargin + 2))
		doc.SetFont(pdfFontFamily, "", 8)
		doc.SetTextColor(110, 110, 110)
		doc.CellFormat(0, 4, fmt.Sprintf("Page %d of {nb}", doc.PageNo()), "", 0, "R", false, 0, "")
		doc.SetTextColor(0, 0, 0)
	})
	doc.AddPage()
	return doc
}

// remaining returns the vertical space left above the bottom margin.
func remaining(doc *fpdf.Fpdf) float64 {
	_, pageH := doc.GetPageSize()
	_, _, _, bottom := doc.GetMargins()
	return pageH - bottom - doc.GetY()
}

func (p providerReport) pdfSection(doc *fpdf.Fpdf, u reportUser, w reportWallet, firstOfUser bool) {
	if remaining(doc) < pdfSectionMinH {
		doc.AddPage()
	} else if doc.GetY() > pdfMargin+9 {
		doc.Ln(4)
	}

	heading := u.Email
	if heading == "" {
		heading = u.UserID
	}
	if firstOfUser {
		doc.Bookmark(heading, 0, -1)
	}
	if w.WalletID != "" {
		doc.Bookmark("Wallet "+w.WalletID, 1, -1)
	}

	// Section heading bar.
	doc.SetFillColor(30, 30, 30)
	doc.SetTextColor(255, 255, 255)
	doc.SetFont(pdfFontFamily, "B", 10)
	doc.CellFormat(0, 7, " "+heading, "", 1, "L", true, 0, "")
	doc.SetTextColor(0, 0, 0)
	doc.Ln(1)

	// Details: label/value pairs in two columns, filled top to bottom so they read in
	// the same order as the CSV columns (user details left, KYC and balance right).
	fields := p.fields(u, w)
	rows := (len(fields) + pdfFieldCols - 1) / pdfFieldCols
	for r := 0; r < rows; r++ {
		for c := 0; c < pdfFieldCols; c++ {
			i := c*rows + r
			if i >= len(fields) {
				break
			}
			doc.SetFont(pdfFontFamily, "B", 8)
			doc.SetTextColor(90, 90, 90)
			doc.CellFormat(pdfLabelW, pdfFieldH, fit(doc, fields[i].label, pdfLabelW), "", 0, "L", false, 0, "")
			doc.SetFont(pdfFontFamily, "", 8)
			doc.SetTextColor(0, 0, 0)
			doc.CellFormat(pdfValueW, pdfFieldH, fit(doc, fields[i].value, pdfValueW), "", 0, "L", false, 0, "")
		}
		doc.Ln(-1)
	}

	// Footnote for the "Local KYC *" label.
	doc.SetFont(pdfFontFamily, "", 7)
	doc.SetTextColor(110, 110, 110)
	doc.CellFormat(0, 4, p.kycDisclaimer(), "", 1, "L", false, 0, "")
	doc.SetTextColor(0, 0, 0)

	if w.Error != "" {
		doc.Ln(1)
		doc.SetFont(pdfFontFamily, "B", 8)
		doc.SetTextColor(180, 0, 0)
		doc.MultiCell(0, pdfFieldH, "Error loading this wallet: "+w.Error, "", "L", false)
		doc.SetTextColor(0, 0, 0)
	}

	doc.Ln(2)
	if w.WalletID == "" {
		p.pdfNote(doc, "No wallet.")
		return
	}
	if len(w.Transactions) == 0 {
		p.pdfNote(doc, "No transactions.")
		return
	}

	pdfTxHeader(doc)
	doc.SetFont(pdfFontFamily, "", pdfTableFontSize)
	for i, t := range w.Transactions {
		// Break pages ourselves so the table header repeats on every page.
		if remaining(doc) < pdfRowH {
			doc.AddPage()
			pdfTxHeader(doc)
			doc.SetFont(pdfFontFamily, "", pdfTableFontSize)
		}
		fill := i%2 == 1
		doc.SetFillColor(245, 245, 245)
		for _, c := range pdfTxColumns {
			doc.CellFormat(c.width, pdfRowH, fit(doc, c.value(t), c.width), "", 0, c.align, fill, 0, "")
		}
		doc.Ln(-1)
	}
}

func (p providerReport) pdfNote(doc *fpdf.Fpdf, text string) {
	doc.SetFont(pdfFontFamily, "", 8)
	doc.SetTextColor(110, 110, 110)
	doc.CellFormat(0, pdfFieldH, text, "", 1, "L", false, 0, "")
	doc.SetTextColor(0, 0, 0)
}

func pdfTxHeader(doc *fpdf.Fpdf) {
	doc.SetFont(pdfFontFamily, "B", pdfTableFontSize)
	doc.SetFillColor(225, 225, 225)
	for _, c := range pdfTxColumns {
		doc.CellFormat(c.width, pdfRowH+1, c.title, "B", 0, c.align, true, 0, "")
	}
	doc.Ln(-1)
}

// fit shortens s with an ellipsis so it fits in a cell of width w at the current font.
// Rows stay one line high; the CSV has the untruncated values.
func fit(doc *fpdf.Fpdf, s string, w float64) string {
	avail := w - 2*doc.GetCellMargin()
	if doc.GetStringWidth(s) <= avail {
		return s
	}
	r := []rune(s)
	for len(r) > 0 && doc.GetStringWidth(string(r)+"…") > avail {
		r = r[:len(r)-1]
	}
	return string(r) + "…"
}
