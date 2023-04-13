package persona

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client interface {
	GetAccount(ctx context.Context, id string) (*AccountData, error)
	CreateInquiry(ctx context.Context, args IndividualAttributes, idempotencyKey string) (*InquiryData, error)
	ResumeInquiry(ctx context.Context, inquiryID, idempotencyKey string) (*InquiryData, error)
}

type client struct {
	baseURL       string
	api           *http.Client
	bearerToken   string
	webhookSecret string
}

func New() Client {
	return &client{
		api:           otelhttp.DefaultClient,
		bearerToken:   os.Getenv("PERSONA_TOKEN"),
		webhookSecret: os.Getenv("PERSONA_WEBHOOK"),
		baseURL:       "https://withpersona.com/api/v1/",
	}
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
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create inquiry")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to resume inquiry")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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

	_, err = c.api.Do(req)
	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("NotImplemented")

}
