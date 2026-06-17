package external

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach-go/crdb/crdbsqlx"
	"github.com/jmoiron/sqlx"
	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/kyc"
	httplog "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client interface {
	CreateSubAccount(ctx context.Context, user user.User, details kyc.IndividualDetails, idNumber, personaInquiryURL string) (*SubAccount, error)
	AddBeneficiary(ctx context.Context, reqStruct CreateBeneficiaryReq) (*AccountBeneficiaries, error)
	ListBeneficiaries(ctx context.Context, limit, page uint) (*ListBeneficiariesResponse, error)
	CreateTransaction(ctx context.Context, amt currency.Amount, idempotencyKey, beneficiaryID, reference string) (string, error)
	ListDeposits(ctx context.Context, page int) ([]Deposit, error)
	GetWithdrawal(ctx context.Context, id string) (*Withdrawal, error)
	TestDeposit(ctx context.Context, reqStruct TestDepositReq) error
	UpdateSubAccount(ctx context.Context, accountID string, reqStruct UpdateSubAccountRequest) error
	BankAccounts(ctx context.Context) (*[]Currency, error)
	GetDeposit(ctx context.Context, id string) (*Deposit, error)
}

// Config holds the configuration for the Xago external client.
type Config struct {
	APIBaseURL      string
	IdentityBaseURL string
	PublicKey       string
	Secret          string
	PolicyID        string
}

type client struct {
	baseURL         string
	identityBaseURL string
	api             *http.Client
	accessToken     AccessToken
	publicKey       string
	secret          string
	policyID        string
	tokenLock       sync.Mutex
	dbc             *sqlx.DB
}

func New(transport *http.Client, dbc *sqlx.DB, cfg Config) Client {
	if transport == nil {
		transport = otelhttp.DefaultClient
	}

	cl := &client{
		dbc:             dbc,
		baseURL:         cfg.APIBaseURL,
		identityBaseURL: cfg.IdentityBaseURL,
		api:             transport,
		accessToken:     AccessToken{},
		publicKey:       cfg.PublicKey,
		secret:          cfg.Secret,
		policyID:        cfg.PolicyID,
	}

	return cl
}

func (c *client) refreshAccessToken(ctx context.Context) error {
	return crdbsqlx.ExecuteTx(ctx, c.dbc, nil, func(tx *sqlx.Tx) error {
		var token AccessToken
		err := tx.GetContext(ctx, &token, "SELECT token, expires_at FROM xago_access_token WHERE id=$1 FOR UPDATE", accessTokenID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		// Another process already updated this, just return the latest.
		if !token.IsExpired() && token.Token != c.accessToken.Token {
			c.accessToken = token
			return nil
		}

		reqUrl, err := url.JoinPath(c.identityBaseURL, "login")
		if err != nil {
			return err
		}

		meta, ok := httplog.MetaForContext(ctx)
		if ok {
			meta.Method = "POST"
			meta.Provider = "xago"
		} else {
			ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
				Method:   "POST",
				Provider: "xago",
			})
		}

		type tokenResp struct {
			Token string `json:"tokenValue"`
		}
		type reqField struct {
			FieldName  string `json:"fieldName"`
			FieldValue string `json:"fieldValue"`
		}
		type reqFormat struct {
			PolicyID string     `json:"policyId"`
			Fields   []reqField `json:"fields"`
		}

		reqStruct := reqFormat{
			PolicyID: c.policyID,
			Fields: []reqField{
				{
					FieldName:  "apiPublicKey",
					FieldValue: c.publicKey,
				},
				{
					FieldName:  "apiSecretKey",
					FieldValue: c.secret,
				},
			},
		}

		reqBody, err := json.Marshal(reqStruct)
		if err != nil {
			return err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewReader(reqBody))
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.api.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to get xargo access token code (%d - %s)", resp.StatusCode, resp.Status)
		}
		
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var respData tokenResp
		err = json.Unmarshal(respBody, &respData)
		if err != nil {
			return err
		}

		c.accessToken = AccessToken{
			Token:     respData.Token,
			ExpiresAt: time.Now().Add(time.Minute * 55),
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO xago_access_token (id, token, expires_at) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET "+
			"token = excluded.token, expires_at = excluded.expires_at ", accessTokenID, c.accessToken.Token, c.accessToken.ExpiresAt)

		return err
	})
}

func (c *client) AccessToken(ctx context.Context, forceRefresh bool) (*AccessToken, error) {
	if !c.accessToken.IsExpired() && !forceRefresh {
		return &c.accessToken, nil
	}
	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()
	if !c.accessToken.IsExpired() && !forceRefresh {
		return &c.accessToken, nil
	}

	err := c.refreshAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	return &c.accessToken, nil
}

func (c *client) BankAccounts(ctx context.Context) (*[]Currency, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "currencies")
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "xago"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "xago",
		})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to retrieve xargo bank account (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respData []Currency
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c *client) CreateSubAccount(ctx context.Context, user user.User, details kyc.IndividualDetails, idNumber, personaInquiryURL string) (*SubAccount, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "company", "accounts")
	if err != nil {
		return nil, err
	}

	reqStruct := SubAccountReq{
		FirstName:       details.FirstName,
		LastName:        details.LastName,
		Email:           user.Email,
		MobileNumber:    user.PhoneNumber,
		IdentityType:    IdentityTypeIndividual,
		PersonaURL:      personaInquiryURL,
		IDNumber:        idNumber,
		PhysicalAddress: details.Address.String(),
	}

	reqBody, err := json.Marshal(reqStruct)
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "xago"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "xago",
		})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	token, err := c.AccessToken(ctx, false)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token, err = c.AccessToken(ctx, true)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)

		resp, err = c.api.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			log.Info("refreshed xago token not authorized for create sub account")
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create xargo sub account (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respData SubAccount
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c *client) AddBeneficiary(ctx context.Context, reqStruct CreateBeneficiaryReq) (*AccountBeneficiaries, error) {
	reqUrl, err := url.JoinPath(c.identityBaseURL, "beneficiaries")
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(reqStruct)
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "xago"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "xago",
		})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	token, err := c.AccessToken(ctx, false)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token, err = c.AccessToken(ctx, true)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)

		resp, err = c.api.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			log.Info("refreshed xago token not authorized for add beneficiary")
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to add xargo beneficiary (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var beneficiary AccountBeneficiaries
	err = json.Unmarshal(respBody, &beneficiary)
	if err != nil {
		return nil, err
	}

	return &beneficiary, nil
}

func (c *client) ListBeneficiaries(ctx context.Context, limit, page uint) (*ListBeneficiariesResponse, error) {
	reqUrl, err := url.Parse(fmt.Sprintf("%s/beneficiaries", c.identityBaseURL))
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "xago"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "xago",
		})
	}

	q := reqUrl.Query()
	q.Set("limit", fmt.Sprintf("%d", limit))
	q.Set("page", fmt.Sprintf("%d", page))
	reqUrl.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	token, err := c.AccessToken(ctx, false)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token, err = c.AccessToken(ctx, true)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)

		resp, err = c.api.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			log.Info("refreshed xago token not authorized for list beneficiaries")
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list xargo beneficiaries (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respData *ListBeneficiariesResponse
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return respData, nil
}

func (c *client) CreateTransaction(ctx context.Context, amt currency.Amount, idempotencyKey, beneficiaryID, reference string) (string, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "transfers")
	if err != nil {
		return "", err
	}
	// TODO: after talking to xago
	if reference == "" {
		reference = "Interledger Wallet"
	}
	reqStruct := CreateTransferReq{
		Amount:        amt.Float64(),
		CurrencyCode:  amt.Currency.String(),
		BeneficiaryID: beneficiaryID,
		Reference:     reference,
	}

	reqBody, err := json.Marshal(reqStruct)
	if err != nil {
		return "", err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "xago"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "xago",
		})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	token, err := c.AccessToken(ctx, false)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := c.api.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token, err = c.AccessToken(ctx, true)
		if err != nil {
			return "", err
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)

		resp, err = c.api.Do(req)
		if err != nil {
			return "", err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			log.Info("refreshed xago token not authorized for create transaction")
		}
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		// Idempotency issue, maybe return something else
		// TODO: lookup transaction
		return "", fmt.Errorf("failed to add xargo transaction, transaction already exists")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to add xargo transaction (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	txID := strings.Replace(string(respBody), "\"", "", -1)

	return txID, nil
}

func (c *client) ListDeposits(ctx context.Context, page int) ([]Deposit, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "company", "transactions")
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "xago"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "xago",
		})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	token, err := c.AccessToken(ctx, false)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Token)

	q := req.URL.Query()
	q.Add("limit", "10")
	q.Add("page", strconv.Itoa(page))
	req.URL.RawQuery = q.Encode()

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token, err = c.AccessToken(ctx, true)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)

		resp, err = c.api.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			log.Info("refreshed xago token not authorized for list deposits")
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list xargo deposits (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var list ListDepositsResponse
	err = json.Unmarshal(respBody, &list)
	if err != nil {
		return nil, err
	}

	return list.Deposits, nil
}

func (c *client) GetDeposit(ctx context.Context, id string) (*Deposit, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "company", "transactions", id)
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "xago"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "xago",
		})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	token, err := c.AccessToken(ctx, false)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token, err = c.AccessToken(ctx, true)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)

		resp, err = c.api.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			log.Info("refreshed xago token not authorized for list deposits")
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list xargo deposits (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var deposit *Deposit
	err = json.Unmarshal(respBody, &deposit)
	if err != nil {
		return nil, err
	}

	return deposit, nil
}

func (c *client) GetWithdrawal(ctx context.Context, id string) (*Withdrawal, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "transactions")
	if err != nil {
		return nil, err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "xago"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "xago",
		})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	token, err := c.AccessToken(ctx, false)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Token)

	q := req.URL.Query()
	q.Add("transactionId", id)
	req.URL.RawQuery = q.Encode()

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token, err = c.AccessToken(ctx, true)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)

		resp, err = c.api.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			log.Info("refreshed xago token not authorized for get withdrawal")
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list xargo deposits (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respData Withdrawal
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

// Only used in local/sandbox environments to simulate a deposit.
func (c *client) TestDeposit(ctx context.Context, reqStruct TestDepositReq) error {
	reqURL, err := url.JoinPath(c.baseURL, "company", "accounts", "testdeposit")
	if err != nil {
		return err
	}

	reqBody, err := json.Marshal(reqStruct)
	if err != nil {
		return err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "xago"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "xago",
		})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	token, err := c.AccessToken(ctx, false)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := c.api.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token, err = c.AccessToken(ctx, true)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)

		resp, err = c.api.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			log.Info("refreshed xago token not authorized for get withdrawal")
		}
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to simulate xago deposits (%d - %s)", resp.StatusCode, resp.Status)
	}

	return err
}

func (c *client) UpdateSubAccount(ctx context.Context, accountID string, reqStruct UpdateSubAccountRequest) error {
	reqURL, err := url.JoinPath(c.baseURL, "company", "accounts", accountID)
	if err != nil {
		return err
	}

	reqBody, err := json.Marshal(reqStruct)
	if err != nil {
		return err
	}

	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "PUT"
		meta.Provider = "xago"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "PUT",
			Provider: "xago",
		})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}
	token, err := c.AccessToken(ctx, false)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token.Token)

	resp, err := c.api.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		token, err = c.AccessToken(ctx, true)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token.Token)

		resp, err = c.api.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode == http.StatusUnauthorized {
			log.Info("refreshed xago token not authorized for add beneficiary")
		}
	}

	if resp.StatusCode != http.StatusOK {
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
				return fmt.Errorf("failed to read body on: update xago inquiry url for (%s) response code: %d", accountID, resp.StatusCode)
		}
		return fmt.Errorf("failed to update xago inquiry url for  (%s), response code: %d error: %v", accountID, resp.StatusCode, respBody)
	}

	return nil
}
