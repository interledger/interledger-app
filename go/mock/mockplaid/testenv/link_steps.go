//go:build e2e
// +build e2e

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (tc *TestContext) mockplaidIsRunning() error {
	resp, err := http.Get(tc.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned %d", resp.StatusCode)
	}
	return nil
}

func (tc *TestContext) iGET(path string) error {
	return tc.do("GET", path, nil)
}

func (tc *TestContext) responseStatusIs(code int) error {
	if tc.lastResponse == nil {
		return fmt.Errorf("no response captured")
	}
	if tc.lastResponse.StatusCode != code {
		return fmt.Errorf("expected status %d, got %d (body: %s)", code, tc.lastResponse.StatusCode, tc.bodyString())
	}
	return nil
}

func (tc *TestContext) responseFieldEquals(name, want string) error {
	got, ok := tc.field(name)
	if !ok {
		return fmt.Errorf("field %q not present in body: %s", name, tc.bodyString())
	}
	if got != want {
		return fmt.Errorf("field %q = %q, want %q", name, got, want)
	}
	return nil
}

func (tc *TestContext) responseFieldPresent(name string) error {
	if _, ok := tc.field(name); !ok {
		return fmt.Errorf("field %q not present in body: %s", name, tc.bodyString())
	}
	return nil
}

func (tc *TestContext) createLinkToken(user string) error {
	body := map[string]interface{}{"user": map[string]string{"client_user_id": user}}
	if err := tc.do("POST", "/link/token/create", body); err != nil {
		return err
	}
	lt, ok := tc.field("link_token")
	if !ok {
		return fmt.Errorf("no link_token in response: %s", tc.bodyString())
	}
	tc.linkToken = lt
	return nil
}

func (tc *TestContext) selectAccount(institutionID, accountKey string) error {
	body := map[string]interface{}{
		"link_token":     tc.linkToken,
		"institution_id": institutionID,
		"account_key":    accountKey,
	}
	if err := tc.do("POST", "/link/session/select", body); err != nil {
		return err
	}
	if tc.lastResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("select failed %d: %s", tc.lastResponse.StatusCode, tc.bodyString())
	}
	var resp struct {
		PublicToken string `json:"public_token"`
		Metadata    struct {
			Accounts []struct {
				ID string `json:"id"`
			} `json:"accounts"`
		} `json:"metadata"`
	}
	if err := json.Unmarshal(tc.lastResponseBody, &resp); err != nil {
		return err
	}
	if len(resp.Metadata.Accounts) != 1 {
		return fmt.Errorf("expected 1 account, got %d", len(resp.Metadata.Accounts))
	}
	tc.publicToken = resp.PublicToken
	tc.prevAccountID = tc.lastAccountID
	tc.lastAccountID = resp.Metadata.Accounts[0].ID
	return nil
}

func (tc *TestContext) accountIDsEqual() error {
	if tc.prevAccountID == "" || tc.lastAccountID == "" {
		return fmt.Errorf("need two selections; prev=%q last=%q", tc.prevAccountID, tc.lastAccountID)
	}
	if tc.prevAccountID != tc.lastAccountID {
		return fmt.Errorf("account ids differ: %q vs %q", tc.prevAccountID, tc.lastAccountID)
	}
	return nil
}

func (tc *TestContext) accountIDsDiffer() error {
	if tc.prevAccountID == "" || tc.lastAccountID == "" {
		return fmt.Errorf("need two selections; prev=%q last=%q", tc.prevAccountID, tc.lastAccountID)
	}
	if tc.prevAccountID == tc.lastAccountID {
		return fmt.Errorf("account ids equal but should differ: %q", tc.lastAccountID)
	}
	return nil
}

func (tc *TestContext) exchangePublicToken() error {
	body := map[string]interface{}{"public_token": tc.publicToken}
	if err := tc.do("POST", "/item/public_token/exchange", body); err != nil {
		return err
	}
	at, ok := tc.field("access_token")
	if !ok {
		return fmt.Errorf("no access_token in response: %s", tc.bodyString())
	}
	tc.accessToken = at
	return nil
}

func (tc *TestContext) resolveInstitution() error {
	// item/get → institution_id
	if err := tc.do("POST", "/item/get", map[string]interface{}{"access_token": tc.accessToken}); err != nil {
		return err
	}
	var itemResp struct {
		Item struct {
			InstitutionID string `json:"institution_id"`
		} `json:"item"`
	}
	if err := json.Unmarshal(tc.lastResponseBody, &itemResp); err != nil {
		return err
	}
	// institutions/get_by_id → name
	if err := tc.do("POST", "/institutions/get_by_id", map[string]interface{}{
		"institution_id": itemResp.Item.InstitutionID,
		"country_codes":  []string{"US"},
	}); err != nil {
		return err
	}
	var instResp struct {
		Institution struct {
			Name string `json:"name"`
		} `json:"institution"`
	}
	if err := json.Unmarshal(tc.lastResponseBody, &instResp); err != nil {
		return err
	}
	tc.institutionName = instResp.Institution.Name
	return nil
}

func (tc *TestContext) institutionNameIs(want string) error {
	if tc.institutionName != want {
		return fmt.Errorf("institution name = %q, want %q", tc.institutionName, want)
	}
	return nil
}
