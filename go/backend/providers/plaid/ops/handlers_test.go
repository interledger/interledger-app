package ops_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gitlab.com/fynbos/backend/providers/plaid/ops"
	"gitlab.com/fynbos/backend/user"
)

// --- fakes -----------------------------------------------------------------

type fakeClient struct {
	createProcessorToken func(ctx context.Context, accessToken, accountID, processor string) (string, error)

	exchangeCalls  int
	processorCalls int
}

func (f *fakeClient) CreateLinkToken(context.Context, string) (string, time.Time, error) {
	return "link-token", time.Now().UTC(), nil
}
func (f *fakeClient) ExchangePublicToken(context.Context, string) (string, string, error) {
	f.exchangeCalls++
	return "access-token", "item-id", nil
}
func (f *fakeClient) CreateProcessorToken(ctx context.Context, accessToken, accountID, processor string) (string, error) {
	f.processorCalls++
	if f.createProcessorToken != nil {
		return f.createProcessorToken(ctx, accessToken, accountID, processor)
	}
	return "processor-token", nil
}

type fakeLinker struct {
	existing      *ops.LinkedIDs
	existingErr   error
	registerErr   error
	registerCalls int
}

func (l *fakeLinker) WithAccountLock(ctx context.Context, _, _ string, fn func(context.Context) error) error {
	return fn(ctx)
}
func (l *fakeLinker) ExistingLink(context.Context, string, string) (*ops.LinkedIDs, error) {
	return l.existing, l.existingErr
}
func (l *fakeLinker) Register(context.Context, ops.LinkPlaidArgs) (*ops.LinkedIDs, error) {
	l.registerCalls++
	if l.registerErr != nil {
		return nil, l.registerErr
	}
	return &ops.LinkedIDs{LinkedAccountID: "la-new", PaymentInformationID: "pi-new"}, nil
}
func (l *fakeLinker) ListLinkedPlaidAccountIDs(context.Context, string) ([]string, error) {
	return nil, nil
}

// --- helpers ---------------------------------------------------------------

func withUser(req *http.Request, id string) *http.Request {
	ctx := context.WithValue(req.Context(), user.CtxKey, &user.User{ID: id})
	return req.WithContext(ctx)
}

// --- tests -----------------------------------------------------------------

func TestCreateLinkToken_Unauthenticated(t *testing.T) {
	h := ops.New(&fakeClient{}, &fakeLinker{}, "fiant")

	rec := httptest.NewRecorder()
	// No user in context.
	h.CreateLinkToken(rec, httptest.NewRequest(http.MethodPost, "/link-token", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestLinkToFiant_IdempotentHit(t *testing.T) {
	client := &fakeClient{}
	linker := &fakeLinker{existing: &ops.LinkedIDs{LinkedAccountID: "la-1", PaymentInformationID: "pi-1"}}
	h := ops.New(client, linker, "fiant")

	rec := httptest.NewRecorder()
	req := withUser(httptest.NewRequest(http.MethodPost, "/link-to-fiant",
		strings.NewReader(`{"public_token":"pub-tok","account_id":"acc-1"}`)), "u1")
	h.LinkToFiant(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	var body struct {
		AlreadyLinked   bool   `json:"already_linked"`
		LinkedAccountID string `json:"linked_account_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !body.AlreadyLinked {
		t.Fatalf("expected already_linked=true")
	}
	// Idempotent hit must NOT exchange, mint, or register.
	if client.exchangeCalls != 0 {
		t.Fatalf("expected no ExchangePublicToken on idempotent hit, got %d", client.exchangeCalls)
	}
	if client.processorCalls != 0 {
		t.Fatalf("expected no processor-token mint on idempotent hit, got %d", client.processorCalls)
	}
	if linker.registerCalls != 0 {
		t.Fatalf("expected no Register on idempotent hit, got %d", linker.registerCalls)
	}
}

func TestLinkToFiant_FreshRegister(t *testing.T) {
	client := &fakeClient{}
	linker := &fakeLinker{existing: nil}
	h := ops.New(client, linker, "fiant")

	rec := httptest.NewRecorder()
	req := withUser(httptest.NewRequest(http.MethodPost, "/link-to-fiant",
		strings.NewReader(`{"public_token":"pub-tok","account_id":"acc-1"}`)), "u1")
	h.LinkToFiant(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}
	if client.exchangeCalls != 1 {
		t.Fatalf("expected exactly one ExchangePublicToken, got %d", client.exchangeCalls)
	}
	if client.processorCalls != 1 {
		t.Fatalf("expected exactly one processor-token mint, got %d", client.processorCalls)
	}
	if linker.registerCalls != 1 {
		t.Fatalf("expected exactly one Register, got %d", linker.registerCalls)
	}
}

func TestLinkToFiant_MintFailure_MapsToBadGateway(t *testing.T) {
	client := &fakeClient{createProcessorToken: func(context.Context, string, string, string) (string, error) {
		return "", errors.New("plaid down")
	}}
	h := ops.New(client, &fakeLinker{}, "fiant")

	rec := httptest.NewRecorder()
	req := withUser(httptest.NewRequest(http.MethodPost, "/link-to-fiant",
		strings.NewReader(`{"public_token":"pub-tok","account_id":"acc-1"}`)), "u1")
	h.LinkToFiant(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 on mint failure, got %d", rec.Code)
	}
}

func TestLinkToFiant_NotConfigured(t *testing.T) {
	// No linker / no processor → 503.
	h := ops.New(&fakeClient{}, nil, "")

	rec := httptest.NewRecorder()
	req := withUser(httptest.NewRequest(http.MethodPost, "/link-to-fiant",
		strings.NewReader(`{"public_token":"pub-tok","account_id":"acc-1"}`)), "u1")
	h.LinkToFiant(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when linker not configured, got %d", rec.Code)
	}
}
