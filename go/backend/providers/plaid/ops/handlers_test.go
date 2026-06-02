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

	"gitlab.com/fynbos/backend/providers/plaid"
	"gitlab.com/fynbos/backend/providers/plaid/ops"
	"gitlab.com/fynbos/backend/user"

	plaidsdk "github.com/plaid/plaid-go/v42/plaid"
)

// --- fakes -----------------------------------------------------------------

type fakeClient struct {
	createProcessorToken func(ctx context.Context, accessToken, accountID, processor string) (string, error)
	removeItemErr        error

	processorCalls int
	removeCalls    int
}

func (f *fakeClient) CreateLinkToken(context.Context, string) (string, time.Time, error) {
	return "link-token", time.Now().UTC(), nil
}
func (f *fakeClient) ExchangePublicToken(context.Context, string) (string, string, error) {
	return "access-token", "item-id", nil
}
func (f *fakeClient) GetInstitutionForItem(context.Context, string) (string, string, error) {
	return "ins_1", "Test Bank", nil
}
func (f *fakeClient) GetAccounts(context.Context, string) (*plaidsdk.AccountsGetResponse, error) {
	return &plaidsdk.AccountsGetResponse{}, nil
}
func (f *fakeClient) GetAuth(context.Context, string) (*plaidsdk.AuthGetResponse, error) {
	return &plaidsdk.AuthGetResponse{}, nil
}
func (f *fakeClient) GetBalance(context.Context, string) (*plaidsdk.AccountsGetResponse, error) {
	return &plaidsdk.AccountsGetResponse{}, nil
}
func (f *fakeClient) GetIdentity(context.Context, string) (*plaidsdk.IdentityGetResponse, error) {
	return &plaidsdk.IdentityGetResponse{}, nil
}
func (f *fakeClient) SyncTransactions(context.Context, string) (*plaid.TransactionsSyncResult, error) {
	return &plaid.TransactionsSyncResult{}, nil
}
func (f *fakeClient) RemoveItem(context.Context, string) error {
	f.removeCalls++
	return f.removeItemErr
}
func (f *fakeClient) CreateProcessorToken(ctx context.Context, accessToken, accountID, processor string) (string, error) {
	f.processorCalls++
	if f.createProcessorToken != nil {
		return f.createProcessorToken(ctx, accessToken, accountID, processor)
	}
	return "processor-token", nil
}

type fakeStore struct {
	set     plaid.TokenSet
	found   bool
	getErr  error
	deleted bool
}

func (s *fakeStore) Get(context.Context, string) (plaid.TokenSet, bool, error) {
	return s.set, s.found, s.getErr
}
func (s *fakeStore) Put(context.Context, string, plaid.TokenSet) error { return nil }
func (s *fakeStore) Delete(context.Context, string) error {
	s.deleted = true
	return nil
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
	h := ops.New(&fakeClient{}, &fakeStore{}, &fakeLinker{}, "fiant")

	rec := httptest.NewRecorder()
	// No user in context.
	h.CreateLinkToken(rec, httptest.NewRequest(http.MethodPost, "/link-token", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestGetState_NotLinked(t *testing.T) {
	h := ops.New(&fakeClient{}, &fakeStore{found: false}, &fakeLinker{}, "fiant")

	rec := httptest.NewRecorder()
	h.GetState(rec, withUser(httptest.NewRequest(http.MethodGet, "/state", nil), "u1"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body plaid.State
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body.Linked {
		t.Fatalf("expected linked=false")
	}
}

func TestGetAccounts_NoLinkedItem(t *testing.T) {
	h := ops.New(&fakeClient{}, &fakeStore{found: false}, &fakeLinker{}, "fiant")

	rec := httptest.NewRecorder()
	h.GetAccounts(rec, withUser(httptest.NewRequest(http.MethodGet, "/accounts", nil), "u1"))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when no item linked, got %d", rec.Code)
	}
}

func TestLinkToFiant_IdempotentHit(t *testing.T) {
	client := &fakeClient{}
	linker := &fakeLinker{existing: &ops.LinkedIDs{LinkedAccountID: "la-1", PaymentInformationID: "pi-1"}}
	store := &fakeStore{found: true, set: plaid.TokenSet{AccessToken: "tok"}}
	h := ops.New(client, store, linker, "fiant")

	rec := httptest.NewRecorder()
	req := withUser(httptest.NewRequest(http.MethodPost, "/link-to-fiant", strings.NewReader(`{"account_id":"acc-1"}`)), "u1")
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
	// Idempotent hit must NOT touch Plaid or Fiant.
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
	store := &fakeStore{found: true, set: plaid.TokenSet{AccessToken: "tok"}}
	h := ops.New(client, store, linker, "fiant")

	rec := httptest.NewRecorder()
	req := withUser(httptest.NewRequest(http.MethodPost, "/link-to-fiant", strings.NewReader(`{"account_id":"acc-1"}`)), "u1")
	h.LinkToFiant(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
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
	store := &fakeStore{found: true, set: plaid.TokenSet{AccessToken: "tok"}}
	h := ops.New(client, store, &fakeLinker{}, "fiant")

	rec := httptest.NewRecorder()
	req := withUser(httptest.NewRequest(http.MethodPost, "/link-to-fiant", strings.NewReader(`{"account_id":"acc-1"}`)), "u1")
	h.LinkToFiant(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 on mint failure, got %d", rec.Code)
	}
}

func TestLinkToFiant_NotConfigured(t *testing.T) {
	// No linker / no processor → 503.
	h := ops.New(&fakeClient{}, &fakeStore{}, nil, "")

	rec := httptest.NewRecorder()
	req := withUser(httptest.NewRequest(http.MethodPost, "/link-to-fiant", strings.NewReader(`{"account_id":"acc-1"}`)), "u1")
	h.LinkToFiant(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when linker not configured, got %d", rec.Code)
	}
}

func TestDisconnect_SoftFailsItemRemove(t *testing.T) {
	client := &fakeClient{removeItemErr: errors.New("item already gone")}
	store := &fakeStore{found: true, set: plaid.TokenSet{AccessToken: "tok", ItemID: "item-1"}}
	h := ops.New(client, store, &fakeLinker{}, "fiant")

	rec := httptest.NewRecorder()
	h.Disconnect(rec, withUser(httptest.NewRequest(http.MethodDelete, "/disconnect", nil), "u1"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 despite Plaid ItemRemove failure, got %d", rec.Code)
	}
	if !store.deleted {
		t.Fatalf("expected local token store entry to be deleted")
	}
	var body struct {
		Disconnected bool `json:"disconnected"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !body.Disconnected {
		t.Fatalf("expected disconnected=true")
	}
}
