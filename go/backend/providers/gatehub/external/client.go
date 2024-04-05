package external

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
	"strings"
	"time"

	"gitlab.com/fynbos/env"
)

var (
	appIDHeader     = "x-gatehub-app-id"
	timestampHeader = "x-gatehub-timestamp"
	signatureHeader = "x-gatehub-signature"
	// managedUserHeader = "x-gatehub-managed-user-uuid"
)

type client struct {
	onOffRampClientID  string
	onboardingClientID string
	exchangeClientID   string
	appID              string
	apiSecret          string
	baseURL            string
	onboardingBaseURL  string
}

func NewClient(appID, secret string) Client {
	onOffRampClientID := "f8119dfd-e563-44ee-9ae2-1e60a4fce74f"
	onboardingClientID := "4df24d1b-5796-4eec-951b-21699d61b970"
	exchangeClientID := "4e28d4df-22d7-414c-97a3-d71956df29ba"
	baseURL := "https://api.sandbox.gatehub.net"
	onboardingBaseURL := "https://onboarding.sandbox.gatehub.net"
	if env.IsProd() {
		onOffRampClientID = "f4c8f30f-7fc3-4aa1-8573-520cb67565e3"
		onboardingClientID = "40a22fc5-9091-4c6f-aff6-a3fddf475b331"
		exchangeClientID = "50e7c590-f6f9-4fa9-9498-260bd978c5d6"
		baseURL = "https://api.gatehub.net"
		onboardingBaseURL = "https://onboarding.gatehub.net"
	}

	return &client{
		onOffRampClientID:  onOffRampClientID,
		onboardingClientID: onboardingClientID,
		exchangeClientID:   exchangeClientID,
		appID:              appID,
		apiSecret:          secret,
		baseURL:            baseURL,
		onboardingBaseURL:  onboardingBaseURL,
	}
}

func (c *client) IssueToken(ctx context.Context, product Product) (*IssueTokenResponse, error) {
	endpoint, err := url.JoinPath(c.baseURL, "auth", "v1", "tokens")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var clientID string
	var payload IssueTokenReqeust
	switch product {
	case Onboarding:
		clientID = c.onboardingClientID
		payload.Scope = []string{"auth", "id", "gatewayapi", "core", "storage"}
	case OnOffRamp:
		clientID = c.onOffRampClientID
		payload.Scope = []string{"auth", "id", "gatewayapi", "core"}
	case Exchange:
		clientID = c.exchangeClientID
		payload.Scope = []string{"auth", "id", "gatewayapi", "core"}
	}
	endpoint = fmt.Sprintf("%s?clientId=%s", endpoint, clientID)

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	err = c.Sign(ctx, req, time.Now(), body, endpoint)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var tokenResp IssueTokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &tokenResp, nil
}

func (c *client) Sign(ctx context.Context, req *http.Request, date time.Time, payload []byte, targetURL string) error {
	base := fmt.Sprintf("%d|%s|%s|%s", date.UnixMilli(), req.Method, targetURL, string(payload))
	base = strings.Trim(base, "|")
	hmac := hmac.New(sha256.New, []byte(c.apiSecret))
	_, err := hmac.Write([]byte(base))
	if err != nil {
		return err
	}

	req.Header.Set(appIDHeader, c.appID)
	req.Header.Set(timestampHeader, fmt.Sprintf("%d", date.UnixMilli()))
	req.Header.Set(signatureHeader, hex.EncodeToString(hmac.Sum(nil)))

	return nil
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
	case http.StatusGone:
		return fmt.Errorf("%w %s, path=%s", ErrGone, string(body), r.Request.URL.Path)
	case http.StatusUnsupportedMediaType:
		return fmt.Errorf("%w %s, path=%s", ErrUnsupportedMediatype, string(body), r.Request.URL.Path)
	case http.StatusMisdirectedRequest:
		return fmt.Errorf("%w %s, path=%s", ErrMisdirectedRequest, string(body), r.Request.URL.Path)
	case http.StatusUnprocessableEntity:
		return fmt.Errorf("%w %s, path=%s", ErrUnprocessableEntity, string(body), r.Request.URL.Path)
	case http.StatusLocked:
		return fmt.Errorf("%w %s, path=%s", ErrLocked, string(body), r.Request.URL.Path)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%w %s, path=%s", ErrTooManyRequests, string(body), r.Request.URL.Path)
	case http.StatusRequestHeaderFieldsTooLarge:
		return fmt.Errorf("%w %s, path=%s", ErrRequestHeadersTooLarge, string(body), r.Request.URL.Path)
	case http.StatusInternalServerError:
		return fmt.Errorf("%w %s, path=%s", ErrServer, string(body), r.Request.URL.Path)
	case http.StatusBadGateway:
		return fmt.Errorf("%w %s, path=%s", ErrBadGateway, string(body), r.Request.URL.Path)
	case http.StatusServiceUnavailable:
		return fmt.Errorf("%w %s, path=%s", ErrServiceUnavailable, string(body), r.Request.URL.Path)
	case http.StatusGatewayTimeout:
		return fmt.Errorf("%w %s, path=%s", ErrGatewayTimeout, string(body), r.Request.URL.Path)
	default:
		return fmt.Errorf("%w Unknown status code=%d, message=%s, path=%s", ErrInternal, r.StatusCode, string(body), r.Request.URL.Path)
	}
}
