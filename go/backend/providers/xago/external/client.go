package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"gitlab.com/fynbos/backend/currency"
	"gitlab.com/fynbos/backend/kyc"
	"gitlab.com/fynbos/backend/user"
	"gitlab.com/fynbos/env"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client interface {
	AccessToken(ctx context.Context) (*AccessToken, error)
	CreateSubAccount(ctx context.Context, user user.User, details kyc.IndividualDetails, idNumbers kyc.PersonaIDNumbers) (*SubAccount, error)
	AddBeneficiary(ctx context.Context, details kyc.IndividualDetails) (*AccountBeneficiaries, error)
	CreateTransaction(ctx context.Context, amt currency.Amount, idempotencyKey, beneficiaryID string) (*Transaction, error)
}

type client struct {
	baseURL     string
	api         *http.Client
	accessToken AccessToken
	publicKey   string
	secret      string
	tokenLock   sync.Mutex
}

func New() Client {
	baseURL := "https://test-api.xago.io:9000/v1"
	if env.IsProd() {
		baseURL = "https://identity-api.xago.io/v1"
	}
	return &client{
		baseURL:     baseURL,
		api:         otelhttp.DefaultClient,
		accessToken: AccessToken{},
		publicKey:   os.Getenv("XARGO_API_PUBLIC_KEY"),
		secret:      os.Getenv("XARGO_API_SECRET"),
	}
}

func (c *client) AccessToken(ctx context.Context) (*AccessToken, error) {
	if !c.accessToken.IsExpired() {
		return &c.accessToken, nil
	}
	c.tokenLock.Lock()
	defer c.tokenLock.Unlock()
	if !c.accessToken.IsExpired() {
		return &c.accessToken, nil
	}

	reqUrl, err := url.JoinPath(c.baseURL, "login")
	if err != nil {
		return nil, err
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
		PolicyID: "TODO",
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
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewReader(reqBody))
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
		return nil, fmt.Errorf("failed to get xargo access token code (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respData tokenResp
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	c.accessToken = AccessToken{
		Token:     respData.Token,
		ExpiresAt: time.Now().Add(time.Hour * 23),
	}

	return &c.accessToken, nil
}

func (c *client) CreateSubAccount(ctx context.Context, user user.User, details kyc.IndividualDetails, idNumbers kyc.PersonaIDNumbers) (*SubAccount, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "company", "users")
	if err != nil {
		return nil, err
	}

	dob, err := strconv.Atoi(details.DateOfBirth.Format("20060102"))
	if err != nil {
		return nil, err
	}

	reqStruct := SubAccountReq{
		FirstName:                  details.FirstName,
		LastName:                   details.LastName,
		Email:                      user.Email,
		MobileNumber:               user.PhoneNumber,
		IdentificationDocumentType: idNumbers.IdentificationNumber,
		IdentificationNumber:       idNumbers.IdentificationClass,
		AddressDocumentType:        "TODO",
		Country:                    idNumbers.IssuingCountry,
		Nationality:                idNumbers.IssuingCountry,
		DateOfBirth:                dob,
		DestinationAddress:         "TODO",
		DestinationTag:             "TODO",
		BeneficiaryAction:          "transit",
	}
	if details.Address != nil {
		reqStruct.Address = details.Address.Line1
		reqStruct.City = details.Address.City
		reqStruct.PostalCode = details.Address.ZipCode
	}

	reqBody, err := json.Marshal(reqStruct)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	token, err := c.AccessToken(ctx)
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

func (c *client) AddBeneficiary(ctx context.Context, details kyc.IndividualDetails) (*AccountBeneficiaries, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "beneficiaries")
	if err != nil {
		return nil, err
	}
	reqStruct := CreateBeneficiaryReq{
		Name:                details.LastName + " " + details.LastName,
		Scope:               "bank",
		CurrencyCode:        "ZAR",
		AccountNumber:       "TODO",
		BranchCode:          "TODO",
		BankName:            "TODO",
		BankCountry:         "TODO",
		AccountName:         details.FirstName,
		BankBeneficiaryType: "IBAN",
		Reference:           details.FirstName + " " + details.LastName[:1],
		Iban:                "TODO",
		Bic:                 "TODO",
		AccountType:         "typeAccountNumber",
	}
	if details.Address != nil {
		reqStruct.BeneficiaryPhysicalAddress = details.Address.Line1
		reqStruct.BeneficiaryCity = details.Address.City
		reqStruct.BeneficiaryCountry = details.Address.CountryCode
		reqStruct.BeneficiaryPostalCode = details.Address.ZipCode
		reqStruct.BeneficiaryAddress = details.Address.Line1
	}

	reqBody, err := json.Marshal(reqStruct)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	token, err := c.AccessToken(ctx)
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to add xargo beneficiary (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respData AccountBeneficiaries
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c *client) CreateTransaction(ctx context.Context, amt currency.Amount, idempotencyKey, beneficiaryID string) (*Transaction, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "transactions", "transfer")
	if err != nil {
		return nil, err
	}
	reqStruct := CreateTransactionReq{
		Values: []TransactionValues{
			{
				Amount:          amt.Float64(),
				CurrencyCode:    amt.Currency.String(),
				BeneficiaryID:   beneficiaryID,
				IdempotencyKey:  idempotencyKey,
				TransactionType: "transfer",
			},
		},
	}

	reqBody, err := json.Marshal(reqStruct)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	token, err := c.AccessToken(ctx)
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

	if resp.StatusCode == http.StatusUnprocessableEntity {
		// Idempotency issue, maybe return something else
		// TODO: lookup transaction
		return nil, fmt.Errorf("failed to add xargo transaction, transaction already exists")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to add xargo transaction (%d - %s)", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var respData Transaction
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}
