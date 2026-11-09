package reports

import (
	"encoding/csv"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestXagoHeaderUnchanged(t *testing.T) {
	want := []string{
		"email", "first_name", "last_name", "user_id", "kratos_country",
		"wallet_id", "wallet_country", "local_kyc", "persona_inquiry_id", "persona_inquiry_state", "xago_linked_account_id", "balance", "available_balance", "balance_currency",
		"transaction_id", "created_at", "updated_at", "type", "state", "provider", "amount", "fee", "currency", "title",
		"error",
	}
	if got := xagoUsersReport.header(); !reflect.DeepEqual(got, want) {
		t.Errorf("xago header changed:\n got %v\nwant %v", got, want)
	}
}

func TestGatehubHeader(t *testing.T) {
	h := gatehubUsersReport.header()
	idx := map[string]int{}
	for i, c := range h {
		idx[c] = i
	}
	for _, c := range []string{"local_kyc", "gatehub_user_id", "gatehub_linked_account_id", "balance"} {
		if _, ok := idx[c]; !ok {
			t.Errorf("gatehub header is missing %q: %v", c, h)
		}
	}
	if idx["gatehub_user_id"] > idx["gatehub_linked_account_id"] {
		t.Errorf("gatehub_user_id should come before gatehub_linked_account_id: %v", h)
	}
}

func readReport(t *testing.T, p providerReport, users []reportUser) (rows [][]string, col func(row []string, name string) string) {
	t.Helper()
	out, err := p.csv(users)
	if err != nil {
		t.Fatal(err)
	}
	rows, err = csv.NewReader(strings.NewReader(out)).ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}
	header := p.header()
	for i, r := range rows {
		if len(r) != len(header) {
			t.Errorf("row %d has %d columns, want %d", i, len(r), len(header))
		}
	}
	col = func(row []string, name string) string {
		for i, h := range header {
			if h == name {
				return row[i]
			}
		}
		t.Fatalf("unknown column %q", name)
		return ""
	}
	return rows, col
}

func TestXagoUsersCSV(t *testing.T) {
	created := time.Date(2026, 10, 1, 12, 0, 0, 0, time.FixedZone("SAST", 2*60*60))
	users := []reportUser{
		{
			UserID: "u1", Email: "a@example.test", FirstName: "Ann", LastName: "Lee", KratosCountry: "ZA",
			Wallets: []reportWallet{
				{
					WalletID: "w1", Country: "ZA", LocalKYC: "Approved", PersonaInquiryID: "inq_1", PersonaInquiryState: "approved",
					LinkedAccountID: "la1", Balance: "10.00", AvailableBalance: "8.00", BalanceCurrency: "ZAR",
					Transactions: []reportTransaction{
						{ID: "t1", CreatedAt: created, UpdatedAt: created, Type: "DEPOSIT", State: "COMPLETED", Provider: "xago", Amount: "10.00", Fee: "0.00", Currency: "ZAR", Title: "Top up, card"},
						{ID: "t2"},
					},
				},
				{WalletID: "w2", Country: "ZA", Error: "boom"},
			},
		},
		{UserID: "u2", Email: "b@example.test", KratosCountry: "ZA"},
	}

	rows, col := readReport(t, xagoUsersReport, users)

	// Header, two transaction rows for w1, one empty row for w2, one row for the wallet-less user.
	if len(rows) != 5 {
		t.Fatalf("got %d rows, want 5", len(rows))
	}
	checks := []struct {
		row        int
		col, value string
	}{
		{1, "local_kyc", "Approved"},
		{1, "persona_inquiry_id", "inq_1"},
		{1, "xago_linked_account_id", "la1"},
		{1, "created_at", "2026-10-01T10:00:00Z"},
		{1, "title", "Top up, card"},
		{2, "created_at", ""},
		{3, "wallet_id", "w2"},
		{3, "error", "boom"},
		{4, "user_id", "u2"},
		{4, "wallet_id", ""},
		{4, "local_kyc", ""},
	}
	for _, c := range checks {
		if got := col(rows[c.row], c.col); got != c.value {
			t.Errorf("row %d %s = %q, want %q", c.row, c.col, got, c.value)
		}
	}
}

func TestGatehubUsersCSV(t *testing.T) {
	users := []reportUser{
		{
			UserID: "u1", Email: "a@example.test", KratosCountry: "DE",
			Wallets: []reportWallet{
				{WalletID: "w1", Country: "DE", LocalKYC: "Approved", ExternalUserID: "gh-1", LinkedAccountID: "la1", Balance: "5.00", AvailableBalance: "5.00", BalanceCurrency: "EUR"},
			},
		},
	}

	rows, col := readReport(t, gatehubUsersReport, users)

	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	for name, want := range map[string]string{
		"gatehub_user_id":           "gh-1",
		"gatehub_linked_account_id": "la1",
		"balance_currency":          "EUR",
		"local_kyc":                 "Approved",
	} {
		if got := col(rows[1], name); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
}

func TestPDF(t *testing.T) {
	now := time.Date(2026, 11, 9, 14, 30, 0, 0, time.UTC)
	var txs []reportTransaction
	for i := 0; i < 120; i++ {
		txs = append(txs, reportTransaction{ID: "tx", CreatedAt: now, Type: "web_monetization_outgoing", Title: strings.Repeat("long title ", 20)})
	}
	users := []reportUser{
		{UserID: "u1", Email: "a@example.test", FirstName: "Łukasz", LastName: "Żółć", Wallets: []reportWallet{
			{WalletID: "w1", Transactions: txs},
			{WalletID: "w2", Error: "boom"},
		}},
		{UserID: "u2", Email: "b@example.test"},
	}

	for _, p := range []providerReport{xagoUsersReport, gatehubUsersReport} {
		out, err := p.pdf(users, now)
		if err != nil {
			t.Fatalf("%s: %v", p.name, err)
		}
		if !strings.HasPrefix(string(out), "%PDF-") {
			t.Fatalf("%s: output is not a PDF", p.name)
		}
		// 120 transactions don't fit on one landscape A4 page.
		if pages := strings.Count(string(out), "/Type /Page\n"); pages < 2 {
			t.Errorf("%s: got %d pages, want the transaction table to span several", p.name, pages)
		}
	}

	if _, err := xagoUsersReport.pdf(nil, now); err != nil {
		t.Errorf("empty report: %v", err)
	}
}

func TestPDFFieldsMatchCSVColumns(t *testing.T) {
	// The PDF's details block follows the CSV column order.
	for _, p := range []providerReport{xagoUsersReport, gatehubUsersReport} {
		var columns []string
		for _, f := range p.fields(reportUser{}, reportWallet{}) {
			columns = append(columns, f.column)
		}
		if want := p.header()[:len(columns)]; !reflect.DeepEqual(columns, want) {
			t.Errorf("%s: pdf fields %v, want csv columns %v", p.name, columns, want)
		}
	}
}

func TestPDFLabelsFit(t *testing.T) {
	// Readable labels must fit their cells untruncated.
	for _, p := range []providerReport{xagoUsersReport, gatehubUsersReport} {
		doc := p.newPDF(time.Now())
		doc.SetFont(pdfFontFamily, "B", 8)
		for _, f := range p.fields(reportUser{}, reportWallet{}) {
			if got := fit(doc, f.label, pdfLabelW); got != f.label {
				t.Errorf("%s: label %q is truncated to %q", p.name, f.label, got)
			}
		}
		doc.SetFont(pdfFontFamily, "B", pdfTableFontSize)
		for _, c := range pdfTxColumns {
			if got := fit(doc, c.title, c.width); got != c.title {
				t.Errorf("table header %q is truncated to %q", c.title, got)
			}
		}
	}
}

func TestFit(t *testing.T) {
	doc := xagoUsersReport.newPDF(time.Now())
	doc.SetFont(pdfFontFamily, "", pdfTableFontSize)
	if got := fit(doc, "short", 50); got != "short" {
		t.Errorf("fit(short) = %q", got)
	}
	got := fit(doc, strings.Repeat("x", 200), 20)
	if !strings.HasSuffix(got, "…") || doc.GetStringWidth(got) > 20 {
		t.Errorf("fit(long) = %q, width %.1f", got, doc.GetStringWidth(got))
	}
}

func TestKYCDisclaimer(t *testing.T) {
	// The disclaimer names the report's provider and fits on one line of the page.
	for _, p := range []providerReport{xagoUsersReport, gatehubUsersReport} {
		d := p.kycDisclaimer()
		if !strings.Contains(d, p.displayName) {
			t.Errorf("%s: disclaimer %q does not name the provider", p.name, d)
		}
		doc := p.newPDF(time.Now())
		doc.SetFont(pdfFontFamily, "", 7)
		if w := doc.GetStringWidth(d); w > 277 {
			t.Errorf("%s: disclaimer is %.0fmm wide, wider than the page", p.name, w)
		}
	}
}
