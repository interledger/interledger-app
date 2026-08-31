package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeIdentity struct {
	email     string
	firstName string
	country   string
}

// fakeKratos serves /admin/identities like Kratos: one page per entry in pages,
// linked by page_token, plus credentials_identifier lookups.
type fakeKratos struct {
	*httptest.Server
	requests atomic.Int64
}

func newFakeKratos(t *testing.T, pages [][]fakeIdentity) *fakeKratos {
	t.Helper()

	byIdentifier := map[string]fakeIdentity{}
	for _, page := range pages {
		for _, id := range page {
			byIdentifier[strings.ToLower(id.email)] = id
		}
	}

	k := &fakeKratos{}
	k.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		k.requests.Add(1)
		// assert, not require: FailNow from a handler goroutine would not fail the test
		assert.Equal(t, "/admin/identities", r.URL.Path)
		query := r.URL.Query()

		if identifier := query.Get("credentials_identifier"); identifier != "" {
			if id, ok := byIdentifier[strings.ToLower(identifier)]; ok {
				writeIdentities(t, w, []fakeIdentity{id})
				return
			}
			writeIdentities(t, w, nil)
			return
		}

		assert.Equal(t, strconv.Itoa(migrationEmailPageSize), query.Get("page_size"))
		page := 0
		if token := query.Get("page_token"); token != "" {
			var err error
			page, err = strconv.Atoi(token)
			assert.NoError(t, err)
		}
		if page >= len(pages) {
			writeIdentities(t, w, nil)
			return
		}
		if page+1 < len(pages) {
			// Kratos sends relative URLs, first and next in one header
			w.Header().Set("Link", fmt.Sprintf(
				`</admin/identities?page_size=500&page_token=0>; rel="first",</admin/identities?page_size=500&page_token=%d>; rel="next"`,
				page+1))
		}
		writeIdentities(t, w, pages[page])
	}))
	t.Cleanup(k.Close)
	return k
}

func writeIdentities(t *testing.T, w http.ResponseWriter, identities []fakeIdentity) {
	t.Helper()
	body := make([]map[string]any, 0, len(identities))
	for i, id := range identities {
		body = append(body, map[string]any{
			"id":         fmt.Sprintf("00000000-0000-0000-0000-%012d", i),
			"schema_id":  "user",
			"schema_url": "http://kratos/schemas/user",
			"traits": map[string]any{
				"email":       id.email,
				"firstName":   id.firstName,
				"countryCode": id.country,
			},
		})
	}
	w.Header().Set("Content-Type", "application/json")
	assert.NoError(t, json.NewEncoder(w).Encode(body))
}

func migrationParams(params SendMigrationEmailParams) SendMigrationEmailParams {
	params.Subject = "Migration"
	params.Paragraphs = []map[string]interface{}{{"paragraph": "We are migrating."}}
	return params
}

func TestListMigrationEmailRecipients(t *testing.T) {
	pages := [][]fakeIdentity{
		{
			{email: "alice@example.com", firstName: "Alice", country: "US"},
			{email: "bruno@example.com", firstName: "Bruno", country: "DE"},
		},
		{
			{email: "carla@example.com", firstName: "Carla", country: "US"},
			{email: "dana@example.com", firstName: "Dana", country: "ZA"},
		},
	}

	activityFor := func(t *testing.T, k *fakeKratos) *Activity {
		t.Helper()
		return &Activity{cfg: Config{KratosURL: k.URL, KratosAdminURL: k.URL}}
	}

	emailsOf := func(recipients []MigrationEmailRecipient) []string {
		out := make([]string, 0, len(recipients))
		for _, r := range recipients {
			out = append(out, r.Email)
		}
		return out
	}

	t.Run("region follows every page", func(t *testing.T) {
		k := newFakeKratos(t, pages)
		got, err := activityFor(t, k).ListMigrationEmailRecipients(context.Background(), migrationParams(SendMigrationEmailParams{Region: "US"}))
		require.NoError(t, err)
		// carla is on page two: without pagination this would only find alice.
		require.ElementsMatch(t, []string{"alice@example.com", "carla@example.com"}, emailsOf(got))
	})

	t.Run("region ALL takes everyone", func(t *testing.T) {
		k := newFakeKratos(t, pages)
		got, err := activityFor(t, k).ListMigrationEmailRecipients(context.Background(), migrationParams(SendMigrationEmailParams{Region: "ALL"}))
		require.NoError(t, err)
		require.ElementsMatch(t,
			[]string{"alice@example.com", "bruno@example.com", "carla@example.com", "dana@example.com"},
			emailsOf(got))
	})

	t.Run("region EU", func(t *testing.T) {
		k := newFakeKratos(t, pages)
		got, err := activityFor(t, k).ListMigrationEmailRecipients(context.Background(), migrationParams(SendMigrationEmailParams{Region: "EU"}))
		require.NoError(t, err)
		require.Equal(t, []MigrationEmailRecipient{{Email: "bruno@example.com", FirstName: "Bruno"}}, got)
	})

	t.Run("region matching nobody is not an error", func(t *testing.T) {
		k := newFakeKratos(t, pages)
		got, err := activityFor(t, k).ListMigrationEmailRecipients(context.Background(), migrationParams(SendMigrationEmailParams{Region: "GB"}))
		require.NoError(t, err)
		require.Empty(t, got)
	})

	t.Run("an identity repeated across pages is sent to once", func(t *testing.T) {
		k := newFakeKratos(t, [][]fakeIdentity{pages[0], {pages[0][0]}})
		got, err := activityFor(t, k).ListMigrationEmailRecipients(context.Background(), migrationParams(SendMigrationEmailParams{Region: "US"}))
		require.NoError(t, err)
		require.Equal(t, []MigrationEmailRecipient{{Email: "alice@example.com", FirstName: "Alice"}}, got)
	})

	t.Run("addresses are looked up directly, not paged through", func(t *testing.T) {
		k := newFakeKratos(t, pages)
		got, err := activityFor(t, k).ListMigrationEmailRecipients(context.Background(),
			migrationParams(SendMigrationEmailParams{Email: "Alice@Example.com, dana@example.com"}))
		require.NoError(t, err)
		require.ElementsMatch(t, []string{"alice@example.com", "dana@example.com"}, emailsOf(got))
		require.Equal(t, int64(2), k.requests.Load(), "one lookup per address")
	})

	t.Run("addresses ignore the region filter", func(t *testing.T) {
		k := newFakeKratos(t, pages)
		got, err := activityFor(t, k).ListMigrationEmailRecipients(context.Background(),
			migrationParams(SendMigrationEmailParams{Region: "EU", Email: "dana@example.com"}))
		require.NoError(t, err)
		require.Equal(t, []MigrationEmailRecipient{{Email: "dana@example.com", FirstName: "Dana"}}, got)
	})

	t.Run("an unknown address fails the run and names it", func(t *testing.T) {
		k := newFakeKratos(t, pages)
		_, err := activityFor(t, k).ListMigrationEmailRecipients(context.Background(),
			migrationParams(SendMigrationEmailParams{Email: "alice@example.com,nobody@example.com,ghost@example.com"}))
		require.EqualError(t, err, "no user found for: ghost@example.com, nobody@example.com")
	})

	t.Run("invalid params are rejected before calling kratos", func(t *testing.T) {
		k := newFakeKratos(t, pages)
		_, err := activityFor(t, k).ListMigrationEmailRecipients(context.Background(), SendMigrationEmailParams{Region: "US"})
		require.EqualError(t, err, "subject is required")
		require.Zero(t, k.requests.Load())
	})

	t.Run("missing kratos config", func(t *testing.T) {
		a := &Activity{}
		_, err := a.ListMigrationEmailRecipients(context.Background(), migrationParams(SendMigrationEmailParams{Region: "US"}))
		require.EqualError(t, err, "kratos URLs are not set")
	})

	t.Run("a kratos error stops the run", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "nope", http.StatusInternalServerError)
		}))
		t.Cleanup(srv.Close)
		a := &Activity{cfg: Config{KratosURL: srv.URL, KratosAdminURL: srv.URL}}
		_, err := a.ListMigrationEmailRecipients(context.Background(), migrationParams(SendMigrationEmailParams{Region: "US"}))
		require.Error(t, err)
	})

	t.Run("endless next links stop at the page cap instead of emailing a partial list", func(t *testing.T) {
		var requests atomic.Int64
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			page := requests.Add(1)
			w.Header().Set("Link", fmt.Sprintf(`</admin/identities?page_token=%d>; rel="next"`, page))
			writeIdentities(t, w, []fakeIdentity{{email: fmt.Sprintf("u%d@example.com", page), firstName: "U", country: "US"}})
		}))
		t.Cleanup(srv.Close)

		a := &Activity{cfg: Config{KratosURL: srv.URL, KratosAdminURL: srv.URL}}
		_, err := a.ListMigrationEmailRecipients(context.Background(), migrationParams(SendMigrationEmailParams{Region: "US"}))
		require.EqualError(t, err, "stopped after 100 pages of identities: refusing to email a partial list")
		require.Equal(t, int64(migrationEmailMaxPages), requests.Load())
	})
}
