package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gitlab.com/fynbos/mock/mockplaid/internal/models"
)

// selectAccount drives a full link-token + select round-trip and returns the
// chosen account_id from the response metadata.
func selectAccount(t *testing.T, h *Handler, instID, key string) string {
	t.Helper()

	// mint link token
	ltReq := httptest.NewRequest(http.MethodPost, "/link/token/create",
		strings.NewReader(`{"user":{"client_user_id":"u_1"}}`))
	ltRR := httptest.NewRecorder()
	h.CreateLinkToken(ltRR, ltReq)
	var lt struct {
		LinkToken string `json:"link_token"`
	}
	_ = json.NewDecoder(ltRR.Body).Decode(&lt)

	body := `{"link_token":"` + lt.LinkToken + `","institution_id":"` + instID + `","account_key":"` + key + `"}`
	req := httptest.NewRequest(http.MethodPost, "/link/session/select", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.SelectAccount(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("select %s/%s: expected 200, got %d (%s)", instID, key, rr.Code, rr.Body.String())
	}
	var resp struct {
		PublicToken string `json:"public_token"`
		Metadata    struct {
			Accounts []struct {
				ID string `json:"id"`
			} `json:"accounts"`
		} `json:"metadata"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode select: %v", err)
	}
	if !strings.HasPrefix(resp.PublicToken, "public-sandbox-") {
		t.Fatalf("bad public_token: %q", resp.PublicToken)
	}
	if len(resp.Metadata.Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(resp.Metadata.Accounts))
	}
	return resp.Metadata.Accounts[0].ID
}

func TestSelect_BankA_StablePerAccount(t *testing.T) {
	h, _ := newTestHandler()
	a1 := selectAccount(t, h, models.InstitutionA, "checking")
	a2 := selectAccount(t, h, models.InstitutionA, "checking")
	if a1 != a2 {
		t.Fatalf("Bank A checking not stable: %q vs %q", a1, a2)
	}
	if a1 != "acc_mock_a_checking" {
		t.Fatalf("unexpected stable id: %q", a1)
	}
	// different account key → different (but stable) id
	s := selectAccount(t, h, models.InstitutionA, "savings")
	if s == a1 {
		t.Fatalf("savings should differ from checking, both %q", s)
	}
}

func TestSelect_BankB_AlwaysNew(t *testing.T) {
	h, _ := newTestHandler()
	b1 := selectAccount(t, h, models.InstitutionB, "checking")
	b2 := selectAccount(t, h, models.InstitutionB, "checking")
	if b1 == b2 {
		t.Fatalf("Bank B should mint fresh ids, got %q twice", b1)
	}
}

func TestSelect_InvalidLinkToken(t *testing.T) {
	h, _ := newTestHandler()
	body := `{"link_token":"nope","institution_id":"ins_mock_a","account_key":"checking"}`
	req := httptest.NewRequest(http.MethodPost, "/link/session/select", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.SelectAccount(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestSelect_ItemPersistedByPublicToken(t *testing.T) {
	h, store := newTestHandler()
	_ = selectAccount(t, h, models.InstitutionA, "checking")
	// the item must be resolvable by its public token for MP9 exchange
	// (we don't have the token here; assert via a fresh select capturing it)
	ltReq := httptest.NewRequest(http.MethodPost, "/link/token/create",
		strings.NewReader(`{"user":{"client_user_id":"u_1"}}`))
	ltRR := httptest.NewRecorder()
	h.CreateLinkToken(ltRR, ltReq)
	var lt struct {
		LinkToken string `json:"link_token"`
	}
	_ = json.NewDecoder(ltRR.Body).Decode(&lt)
	body := `{"link_token":"` + lt.LinkToken + `","institution_id":"ins_mock_b","account_key":"checking"}`
	req := httptest.NewRequest(http.MethodPost, "/link/session/select", strings.NewReader(body))
	rr := httptest.NewRecorder()
	h.SelectAccount(rr, req)
	var resp struct {
		PublicToken string `json:"public_token"`
	}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	item, err := store.GetItemByPublicToken(context.Background(), resp.PublicToken)
	if err != nil {
		t.Fatalf("item not stored by public token: %v", err)
	}
	if item.AccessToken == "" || item.InstitutionName != "Platypus Bank" {
		t.Fatalf("unexpected stored item: %+v", item)
	}
}
