package persona

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client interface {
	GetAccount(ctx context.Context, id string) (*AccountData, error)
	CreateInquiry(ctx context.Context, args IndividualAttributes, idempotencyKey string) (*InquiryData, error)
	ResumeInquiry(ctx context.Context, inquiryID, idempotencyKey string) (*InquiryData, error)
	LookupInquiry(ctx context.Context, inquiryID string) (*Inquiry, error)
	ValidateWebhook(req *http.Request, body []byte) bool
	RemoveTag(ctx context.Context, accountID string, tag string) (*AccountData, error)
}

type client struct {
	baseURL       string
	api           *http.Client
	bearerToken   string
	webhookSecret string
}

func New() Client {
	baseURL := "https://api.withpersona.com/api/v1/"
	if override := os.Getenv("PERSONA_BASE_URL"); override != "" {
		baseURL = override
	}

	return &client{
		api:           otelhttp.DefaultClient,
		bearerToken:   os.Getenv("PERSONA_TOKEN"),
		webhookSecret: os.Getenv("PERSONA_WEBHOOK_TOKEN"),
		baseURL:       baseURL,
	}
}

func (c *client) LookupInquiry(ctx context.Context, inquiryID string) (*Inquiry, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "inquiries", inquiryID)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Persona-Version", "2023-01-05")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to resume inquiry status (%s) body (%s)", resp.Status, string(respBody))
	}

	var respData Inquiry
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData, nil
}

func (c *client) CreateInquiry(ctx context.Context, args IndividualAttributes, idempotencyKey string) (*InquiryData, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "inquiries")
	if err != nil {
		return nil, err
	}

	reqBody := CreateInquiryReq{
		Data: CreateInquiryReqData{Attributes: args},
	}

	reqJson, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewReader(reqJson))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Persona-Version", "2023-01-05")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		// check if the body contains the idempotency error
		if resp.StatusCode == http.StatusBadRequest {
			var respData ErrorResponse
			err = json.Unmarshal(respBody, &respData)
			if err == nil {
				if strings.Contains(respData.Errors[0].Title, "Idempotency") {
					return nil, ErrIdempotencyDuplicate
				}
			}
		}
		return nil, fmt.Errorf("failed to create inquiry status (%s) body (%s)", resp.Status, string(respBody))
	}

	var respData Inquiry
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData.Data, nil
}

func (c *client) ResumeInquiry(ctx context.Context, inquiryID, idempotencyKey string) (*InquiryData, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "inquiries", inquiryID, "resume")
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Persona-Version", "2023-01-05")
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to resume inquiry status (%s) body (%s)", resp.Status, string(respBody))
	}

	var respData Inquiry
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData.Data, nil
}

func (c *client) GetAccount(ctx context.Context, id string) (*AccountData, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "accounts", id)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Persona-Version", "2023-01-05")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get account status (%s) body(%s)", resp.Status, string(respBody))
	}

	var respData Account
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData.Data, nil
}

func (c *client) ValidateWebhook(req *http.Request, body []byte) bool {
	sigs := strings.Split(req.Header.Get("Persona-Signature"), ",")
	if len(sigs) != 2 {
		return false
	}
	var t, v1 string
	for i, v := range sigs {
		if i == 0 {
			t = v[2:]
		}
		if i == 1 {
			v1 = v[3:]
		}
	}

	mac := hmac.New(sha256.New, []byte(c.webhookSecret))
	mac.Write(append([]byte(t+"."), body...))
	expectedMac := hex.EncodeToString(mac.Sum(nil))

	return strings.EqualFold(expectedMac, v1)
}

func (c *client) RemoveTag(ctx context.Context, accountID, tag string) (*AccountData, error) {
	reqUrl, err := url.JoinPath(c.baseURL, "accounts", accountID, "remove-tag")
	if err != nil {
		return nil, err
	}

	reqBody := map[string]interface{}{
		"meta": map[string]string{
			"tag_name": tag,
		},
	}

	reqJson, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqUrl, bytes.NewReader(reqJson))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.bearerToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Persona-Version", "2023-01-05")

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to remove tag on account. status=(%s) body=(%s)", resp.Status, string(respBody))
	}

	var respData Account
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return nil, err
	}

	return &respData.Data, nil
}
