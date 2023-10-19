package smileid

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"gitlab.com/fynbos/env"
)

const prodEndpoint = "https://api.smileidentity.com/v1"
const sandboxEndpoint = "https://testapi.smileidentity.com/v1"

type Client interface {
	GetToken(ctx context.Context, userID, jobID string, product Product) (string, error)
}

type client struct {
	baseURL    string
	partnerID  string
	apiKey     string
	webhookURL string
}

func New(partnerID, apiKey string) Client {
	baseURL := prodEndpoint
	if !env.IsProd() {
		baseURL = sandboxEndpoint
	}

	return &client{
		baseURL:    baseURL,
		partnerID:  partnerID,
		apiKey:     apiKey,
		webhookURL: "https://eu1.fynbos.dev/webhooks/smileid",
	}
}

func (c client) GetToken(ctx context.Context, userID, jobID string, product Product) (string, error) {
	endpoint, err := url.JoinPath(c.baseURL, "token")
	if err != nil {
		return "", err
	}

	timestamp := time.Now()
	signature, err := GenerateSignature(ctx, c.partnerID, c.apiKey, timestamp)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload := struct {
		UserID      string `json:"user_id"`
		JobID       string `json:"job_id"`
		CallbackURL string `json:"callback_url"`
		PartnerID   string `json:"partner_id"`
		Signature   string `json:"signature"`
		Timestamp   string `json:"timestamp"`
		Product     string `json:"product"`
	}{
		UserID:      userID,
		JobID:       jobID,
		PartnerID:   c.partnerID,
		CallbackURL: c.webhookURL,
		Signature:   signature,
		Timestamp:   timestamp.Format(time.RFC3339),
		Product:     string(product),
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Set("content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	err = checkResponseStatusCode(resp)
	if err != nil {
		return "", err
	}

	var tokenResponse GetTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return tokenResponse.Token, nil
}

func GenerateSignature(ctx context.Context, partnerID, apiKey string, timestamp time.Time) (string, error) {
	hasher := hmac.New(sha256.New, []byte(apiKey))
	_, err := hasher.Write([]byte(timestamp.Format(time.RFC3339)))
	if err != nil {
		return "", err
	}

	_, err = hasher.Write([]byte(partnerID))
	if err != nil {
		return "", err
	}

	_, err = hasher.Write([]byte("sid_request"))
	if err != nil {
		return "", err
	}

	signature := hasher.Sum(nil)

	return base64.StdEncoding.EncodeToString(signature), nil
}

func checkResponseStatusCode(r *http.Response) error {
	if http.StatusOK <= r.StatusCode && r.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}
	origRespBody := make([]byte, len(body))
	copy(origRespBody, body)
	defer func() {
		if body != nil {
			r.Body = io.NopCloser(bytes.NewBuffer(origRespBody))
		}
	}()

	switch r.StatusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("%w %s, path=%s", ErrBadRequest, string(body), r.Request.URL.Path)
	case http.StatusUnauthorized:
		return fmt.Errorf("%w %s, path=%s", ErrUnauthorized, string(body), r.Request.URL.Path)
	case http.StatusForbidden:
		return fmt.Errorf("%w %s, path=%s", ErrForbidden, string(body), r.Request.URL.Path)
	case http.StatusNotFound:
		return fmt.Errorf("%w %s, path=%s", ErrNotFound, string(body), r.Request.URL.Path)
	case http.StatusMethodNotAllowed:
		return fmt.Errorf("%w %s, path=%s", ErrMethodNotAllowed, string(body), r.Request.URL.Path)
	case http.StatusNotAcceptable:
		return fmt.Errorf("%w %s, path=%s", ErrNotAcceptable, string(body), r.Request.URL.Path)
	case http.StatusConflict:
		return fmt.Errorf("%w %s, path=%s", ErrConflict, string(body), r.Request.URL.Path)
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("%w %s, path=%s", ErrUnprocessableEntity, string(body), r.Request.URL.Path)
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusGatewayTimeout:
		return fmt.Errorf("%w %s, path=%s", ErrServer, string(body), r.Request.URL.Path)
	case http.StatusServiceUnavailable:
		return fmt.Errorf("%w %s, path=%s", ErrServiceUnavailable, string(body), r.Request.URL.Path)
	default:
		return fmt.Errorf("%w Unknown status code=%d, message=%s, path=%s", ErrInternal, r.StatusCode, string(body), r.Request.URL.Path)
	}

}
