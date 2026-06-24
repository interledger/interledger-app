package sendgrid

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sg "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func newTestClient(baseURL string) *client {
	mailer := sg.NewSendClient("test-api-key")
	mailer.BaseURL = baseURL

	return &client{
		from:   mail.NewEmail("Interledger", "support@interledger.app"),
		mailer: mailer,
	}
}

func TestSendTemplateReturnsErrorForNon2xxStatus(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("invalid api key"))
	}))
	defer ts.Close()

	c := newTestClient(ts.URL)
	err := c.SendTemplate(
		context.Background(),
		"Subject",
		[]Email{{Name: "Alice", Address: "alice@example.com"}},
		"d-template",
		map[string]interface{}{"firstName": "Alice"},
		nil,
	)
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "status=401") {
		t.Fatalf("expected status code in error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "invalid api key") {
		t.Fatalf("expected response body in error, got: %s", errMsg)
	}
}

func TestSendTemplateSucceedsFor2xxStatus(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("x-message-id", "msg-123")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("accepted"))
	}))
	defer ts.Close()

	c := newTestClient(ts.URL)
	err := c.SendTemplate(
		context.Background(),
		"Subject",
		[]Email{{Name: "Bob", Address: "bob@example.com"}},
		"d-template",
		map[string]interface{}{"firstName": "Bob"},
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error for 2xx status, got: %v", err)
	}
}

func TestFirstHeaderValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		headers map[string][]string
		key     string
		want    string
	}{
		{
			name: "case insensitive match",
			headers: map[string][]string{
				"x-message-id": {" msg-001 "},
			},
			key:  "X-Message-Id",
			want: "msg-001",
		},
		{
			name: "missing key",
			headers: map[string][]string{
				"X-Other": {"abc"},
			},
			key:  "X-Message-Id",
			want: "",
		},
		{
			name: "empty value list",
			headers: map[string][]string{
				"X-Message-Id": {},
			},
			key:  "X-Message-Id",
			want: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := firstHeaderValue(tt.headers, tt.key)
			if got != tt.want {
				t.Fatalf("firstHeaderValue(...) = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMaskEmailAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "standard", in: "abcdef@example.com", want: "abc***@example.com"},
		{name: "short local", in: "ab@example.com", want: "ab***@example.com"},
		{name: "invalid address", in: "abcdef", want: "abc***"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := maskEmailAddress(tt.in)
			if got != tt.want {
				t.Fatalf("maskEmailAddress(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
