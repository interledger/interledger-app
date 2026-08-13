package jobs

import (
	"net/http"
	"testing"

	"github.com/interledger/interledger-app/go/backend/country"
	"github.com/stretchr/testify/require"
)

func TestValidateSendMigrationEmailParams(t *testing.T) {
	paragraphs := []map[string]interface{}{{"paragraph": "Hello"}}

	t.Run("requires subject", func(t *testing.T) {
		err := validateSendMigrationEmailParams(SendMigrationEmailParams{
			Paragraphs: paragraphs,
			Region:     "US",
		})
		require.EqualError(t, err, "subject is required")
	})

	t.Run("requires paragraphs", func(t *testing.T) {
		err := validateSendMigrationEmailParams(SendMigrationEmailParams{
			Subject: "Migration",
			Region:  "US",
		})
		require.EqualError(t, err, "paragraphs are required")
	})

	t.Run("requires region when email empty", func(t *testing.T) {
		err := validateSendMigrationEmailParams(SendMigrationEmailParams{
			Subject:    "Migration",
			Paragraphs: paragraphs,
		})
		require.EqualError(t, err, "region is required when email is not set")
	})

	t.Run("allows email without region", func(t *testing.T) {
		err := validateSendMigrationEmailParams(SendMigrationEmailParams{
			Subject:    "Migration",
			Paragraphs: paragraphs,
			Email:      "test@example.com",
		})
		require.NoError(t, err)
	})

	t.Run("allows several emails without region", func(t *testing.T) {
		err := validateSendMigrationEmailParams(SendMigrationEmailParams{
			Subject:    "Migration",
			Paragraphs: paragraphs,
			Email:      "a@example.com, b@example.com",
		})
		require.NoError(t, err)
	})

	t.Run("requires region when email is only separators", func(t *testing.T) {
		err := validateSendMigrationEmailParams(SendMigrationEmailParams{
			Subject:    "Migration",
			Paragraphs: paragraphs,
			Email:      " , ",
		})
		require.EqualError(t, err, "region is required when email is not set")
	})

	t.Run("allows region ALL", func(t *testing.T) {
		err := validateSendMigrationEmailParams(SendMigrationEmailParams{
			Subject:    "Migration",
			Paragraphs: paragraphs,
			Region:     "all",
		})
		require.NoError(t, err)
	})

	t.Run("rejects unknown region", func(t *testing.T) {
		err := validateSendMigrationEmailParams(SendMigrationEmailParams{
			Subject:    "Migration",
			Paragraphs: paragraphs,
			Region:     "NARNIA",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown region")
	})
}

func TestResolveMigrationCountries(t *testing.T) {
	t.Run("EU expands to EU countries", func(t *testing.T) {
		got, err := resolveMigrationCountries("eu")
		require.NoError(t, err)
		require.True(t, got[country.DE])
		require.True(t, got[country.FR])
		require.False(t, got[country.US])
		require.Equal(t, len(country.EUCountries), len(got))
	})

	t.Run("single country", func(t *testing.T) {
		got, err := resolveMigrationCountries("ZA")
		require.NoError(t, err)
		require.Equal(t, map[country.Country]bool{country.ZA: true}, got)
	})

	t.Run("ALL means no country filter", func(t *testing.T) {
		got, err := resolveMigrationCountries("all")
		require.NoError(t, err)
		require.Nil(t, got)
	})
}

func TestMigrationRecipientFromIdentity(t *testing.T) {
	usCountries, err := resolveMigrationCountries("US")
	require.NoError(t, err)
	euCountries, err := resolveMigrationCountries("EU")
	require.NoError(t, err)

	traits := map[string]any{
		"email":       "alice@example.com",
		"firstName":   "Alice",
		"countryCode": "US",
	}

	t.Run("matches region", func(t *testing.T) {
		got, ok := migrationRecipientFromIdentity(traits, nil, usCountries)
		require.True(t, ok)
		require.Equal(t, MigrationEmailRecipient{Email: "alice@example.com", FirstName: "Alice"}, got)
	})

	t.Run("skips non matching region", func(t *testing.T) {
		_, ok := migrationRecipientFromIdentity(traits, nil, euCountries)
		require.False(t, ok)
	})

	t.Run("matches single email case insensitive", func(t *testing.T) {
		got, ok := migrationRecipientFromIdentity(traits, parseMigrationEmails("Alice@Example.com"), nil)
		require.True(t, ok)
		require.Equal(t, "alice@example.com", got.Email)
	})

	t.Run("matches one of several addresses", func(t *testing.T) {
		got, ok := migrationRecipientFromIdentity(traits, parseMigrationEmails("bob@example.com, alice@example.com"), nil)
		require.True(t, ok)
		require.Equal(t, "alice@example.com", got.Email)
	})

	t.Run("skips other email in test mode", func(t *testing.T) {
		_, ok := migrationRecipientFromIdentity(traits, parseMigrationEmails("bob@example.com"), nil)
		require.False(t, ok)
	})

	t.Run("addresses win over the country filter", func(t *testing.T) {
		got, ok := migrationRecipientFromIdentity(traits, parseMigrationEmails("alice@example.com"), euCountries)
		require.True(t, ok)
		require.Equal(t, "alice@example.com", got.Email)
	})

	t.Run("skips missing email", func(t *testing.T) {
		_, ok := migrationRecipientFromIdentity(map[string]any{"firstName": "Alice", "countryCode": "US"}, nil, usCountries)
		require.False(t, ok)
	})

	t.Run("skips traits that are not an object", func(t *testing.T) {
		_, ok := migrationRecipientFromIdentity("alice@example.com", nil, usCountries)
		require.False(t, ok)
	})

	t.Run("nil countries matches every country", func(t *testing.T) {
		allCountries, err := resolveMigrationCountries("ALL")
		require.NoError(t, err)

		for _, code := range []string{"US", "DE", "ZA", ""} {
			got, ok := migrationRecipientFromIdentity(map[string]any{
				"email":       "alice@example.com",
				"firstName":   "Alice",
				"countryCode": code,
			}, nil, allCountries)
			require.True(t, ok, "country %q", code)
			require.Equal(t, "alice@example.com", got.Email)
		}
	})
}

func TestParseMigrationEmails(t *testing.T) {
	require.Equal(t, map[string]bool{"a@example.com": true, "b@example.com": true},
		parseMigrationEmails(" A@Example.com , b@example.com ,, "))
	require.Empty(t, parseMigrationEmails(""))
	require.Empty(t, parseMigrationEmails(" , "))
	require.Equal(t, map[string]bool{"a@example.com": true},
		parseMigrationEmails("a@example.com,A@EXAMPLE.COM"))
}

func TestNextPageToken(t *testing.T) {
	linkHeader := func(v string) *http.Response {
		return &http.Response{Header: http.Header{"Link": []string{v}}}
	}

	t.Run("reads token from next link", func(t *testing.T) {
		resp := linkHeader(`</admin/identities?page_size=500&page_token=eyJvZmZzZXQiOiI1MDAifQ>; rel="next",` +
			`</admin/identities?page_size=500&page_token=1>; rel="first"`)
		require.Equal(t, "eyJvZmZzZXQiOiI1MDAifQ", nextPageToken(resp))
	})

	t.Run("no next link", func(t *testing.T) {
		require.Empty(t, nextPageToken(linkHeader(`</admin/identities?page_size=500&page_token=1>; rel="first"`)))
	})

	t.Run("no link header", func(t *testing.T) {
		require.Empty(t, nextPageToken(&http.Response{Header: http.Header{}}))
	})

	t.Run("nil response", func(t *testing.T) {
		require.Empty(t, nextPageToken(nil))
	})
}
