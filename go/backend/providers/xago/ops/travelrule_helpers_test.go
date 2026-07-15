package ops

import (
	"bytes"
	"encoding/csv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinNonEmpty(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		sep   string
		parts []string
		out   string
	}{
		{
			name:  "all empty",
			sep:   ", ",
			parts: []string{"", "", ""},
			out:   "",
		},
		{
			name:  "empties skipped",
			sep:   ", ",
			parts: []string{"", "Cape Town", "", "ZA"},
			out:   "Cape Town, ZA",
		},
		{
			name:  "all present",
			sep:   " ",
			parts: []string{"John", "Doe"},
			out:   "John Doe",
		},
		{
			name:  "single value",
			sep:   ", ",
			parts: []string{"only"},
			out:   "only",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.out, joinNonEmpty(tc.sep, tc.parts...))
		})
	}
}

func TestFormatDateOfBirth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name             string
		year, month, day int
		out              string
	}{
		{name: "valid", year: 1990, month: 12, day: 31, out: "1990-12-31"},
		{name: "zero padded", year: 3, month: 2, day: 5, out: "0003-02-05"},
		{name: "missing year", year: 0, month: 2, day: 5, out: ""},
		{name: "missing month", year: 1990, month: 0, day: 5, out: ""},
		{name: "missing day", year: 1990, month: 2, day: 0, out: ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.out, formatDateOfBirth(tc.year, tc.month, tc.day))
		})
	}
}

func TestTravelRuleBatchTotal(t *testing.T) {
	t.Parallel()

	const size = 10

	cases := []struct {
		name string
		n    int
		out  int
	}{
		{name: "empty", n: 0, out: 0},
		{name: "single", n: 1, out: 1},
		{name: "exactly one batch", n: size, out: 1},
		{name: "one over a batch", n: size + 1, out: 2},
		{name: "exactly two batches", n: 2 * size, out: 2},
		{name: "one over two batches", n: 2*size + 1, out: 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.out, travelRuleBatchTotal(tc.n, size))
		})
	}
}

func TestBuildTravelRuleCSV(t *testing.T) {
	t.Parallel()

	wantHeader := []string{
		"transaction_reference",
		"originator_name",
		"originator_account_id",
		"originator_address",
		"originator_place_of_birth",
		"originator_date_of_birth",
		"beneficiary_name",
		"beneficiary_account_id",
	}

	t.Run("header only when empty", func(t *testing.T) {
		t.Parallel()

		out, err := buildTravelRuleCSV(nil)
		require.NoError(t, err)

		rows, err := csv.NewReader(bytes.NewReader(out)).ReadAll()
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, wantHeader, rows[0])
	})

	t.Run("row maps in column order", func(t *testing.T) {
		t.Parallel()

		in := travelRuleReportRow{
			TransactionReference:   "pay-1",
			OriginatorName:         "John Doe",
			OriginatorAccountID:    "gh-uuid",
			OriginatorAddress:      "1 Main St, Cape Town, ZA",
			OriginatorPlaceOfBirth: "Cape Town, ZA",
			OriginatorDateOfBirth:  "1990-12-31",
			BeneficiaryName:        "Jane Roe",
			BeneficiaryAccountID:   "xago-acc",
		}

		out, err := buildTravelRuleCSV([]travelRuleReportRow{in})
		require.NoError(t, err)

		rows, err := csv.NewReader(bytes.NewReader(out)).ReadAll()
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Equal(t, wantHeader, rows[0])
		assert.Equal(t, []string{
			"pay-1",
			"John Doe",
			"gh-uuid",
			"1 Main St, Cape Town, ZA",
			"Cape Town, ZA",
			"1990-12-31",
			"Jane Roe",
			"xago-acc",
		}, rows[1])
	})

	t.Run("fields with commas quotes and newlines round-trip", func(t *testing.T) {
		t.Parallel()

		in := travelRuleReportRow{
			TransactionReference: "pay-2",
			OriginatorName:       `Doe, "Johnny"`,
			OriginatorAddress:    "line1\nline2",
			BeneficiaryName:      "Jane Roe",
		}

		out, err := buildTravelRuleCSV([]travelRuleReportRow{in})
		require.NoError(t, err)

		rows, err := csv.NewReader(bytes.NewReader(out)).ReadAll()
		require.NoError(t, err)
		require.Len(t, rows, 2)
		assert.Equal(t, `Doe, "Johnny"`, rows[1][1])
		assert.Equal(t, "line1\nline2", rows[1][3])
	})
}
