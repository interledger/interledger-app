package reports

import (
	"encoding/csv"
	"strings"
	"testing"
	"time"
)

func TestXagoUsersCSV(t *testing.T) {
	created := time.Date(2026, 10, 1, 12, 0, 0, 0, time.FixedZone("SAST", 2*60*60))
	users := []xagoReportUser{
		{
			UserID: "u1", Email: "a@example.test", FirstName: "Ann", LastName: "Lee", KratosCountry: "ZA",
			Wallets: []xagoReportWallet{
				{
					WalletID: "w1", Country: "ZA", PersonaInquiryID: "inq_1", PersonaInquiryState: "approved",
					XagoLinkedAccountID: "la1", Balance: "10.00", AvailableBalance: "8.00", BalanceCurrency: "ZAR",
					Transactions: []xagoReportTransaction{
						{ID: "t1", CreatedAt: created, UpdatedAt: created, Type: "DEPOSIT", State: "COMPLETED", Provider: "xago", Amount: "10.00", Fee: "0.00", Currency: "ZAR", Title: "Top up, card"},
						{ID: "t2"},
					},
				},
				{WalletID: "w2", Country: "ZA", Error: "boom"},
			},
		},
		{UserID: "u2", Email: "b@example.test", KratosCountry: "ZA"},
	}

	out, err := xagoUsersCSV(users)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(strings.NewReader(out)).ReadAll()
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	// Header, two transaction rows for w1, one empty row for w2, one row for the wallet-less user.
	if len(rows) != 5 {
		t.Fatalf("got %d rows, want 5:\n%s", len(rows), out)
	}
	for i, r := range rows {
		if len(r) != len(xagoUsersCSVHeader) {
			t.Errorf("row %d has %d columns, want %d", i, len(r), len(xagoUsersCSVHeader))
		}
	}

	col := func(row []string, name string) string {
		for i, h := range xagoUsersCSVHeader {
			if h == name {
				return row[i]
			}
		}
		t.Fatalf("unknown column %q", name)
		return ""
	}

	if got := col(rows[1], "persona_inquiry_id"); got != "inq_1" {
		t.Errorf("persona_inquiry_id = %q, want inq_1", got)
	}
	if got := col(rows[1], "created_at"); got != "2026-10-01T10:00:00Z" {
		t.Errorf("created_at = %q, want UTC RFC3339", got)
	}
	if got := col(rows[1], "title"); got != "Top up, card" {
		t.Errorf("title = %q, want comma preserved", got)
	}
	if got := col(rows[2], "created_at"); got != "" {
		t.Errorf("zero created_at = %q, want empty", got)
	}
	if got := col(rows[3], "wallet_id"); got != "w2" {
		t.Errorf("wallet_id = %q, want w2", got)
	}
	if got := col(rows[3], "error"); got != "boom" {
		t.Errorf("error = %q, want boom", got)
	}
	if got := col(rows[4], "user_id"); got != "u2" {
		t.Errorf("user_id = %q, want u2", got)
	}
	if got := col(rows[4], "wallet_id"); got != "" {
		t.Errorf("wallet-less user wallet_id = %q, want empty", got)
	}
}
