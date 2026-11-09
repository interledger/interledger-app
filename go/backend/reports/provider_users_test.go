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
