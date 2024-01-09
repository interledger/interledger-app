package external

import (
	"bytes"
	"context"
	"crypto"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jws"
	httplog "gitlab.com/fynbos/backend/providers/http"
	"gitlab.com/fynbos/env"
	"gitlab.com/fynbos/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var (
	ptiClientIDHeader  = "x-pti-client-id"
	ptiSignatureHeader = "x-pti-signature"
)

type client struct {
	baseURL    string
	clientID   string
	privateKey jwk.Key

	// Thumbprint is used as the `kid` field in the jwt protected header
	publicKeyThumbprint string
	api                 *http.Client
}

var _ Client = client{}

type ClientArgs struct {
	ClientID   string
	PrivateKey jwk.Key
	Transport  *http.Client
	BaseURL    string
}

func New(args ClientArgs) Client {
	base := "https://pti.apistaging.pticlient.com/v1"
	if args.BaseURL != "" {
		base = args.BaseURL
	} else if env.IsLocal() {
		base = "http://pti.mock"
	}

	transport := otelhttp.DefaultClient
	if args.Transport != nil {
		transport = args.Transport
	}

	// thumbprint is used as the `kid` field in the jwt protected header
	var thumbprint []byte
	if args.PrivateKey != nil {
		publicKey, err := args.PrivateKey.PublicKey()
		if err != nil {
			log.Fatalln(fmt.Errorf("%w %s", ErrInternal, err))
		}

		thumbprint, err = publicKey.Thumbprint(crypto.SHA256)
		if err != nil {
			log.Fatalln(fmt.Errorf("%w %s", ErrInternal, err))
		}

		// Remove `kid` otherwise lestrat-jws will not let us override the field in the protected header.
		err = args.PrivateKey.Remove("kid")
		if err != nil {
			log.Fatalln(fmt.Errorf("%w Failed to remove `kid` field from jwk %s", ErrInternal, err))
		}
	}

	return &client{
		baseURL:             base,
		clientID:            args.ClientID,
		privateKey:          args.PrivateKey,
		publicKeyThumbprint: base64.RawURLEncoding.EncodeToString(thumbprint),
		api:                 transport,
	}
}

func Sign(ctx context.Context, r *http.Request, date time.Time, payload []byte, key crypto.PrivateKey, publicKeyThumbprint string) error {
	var contentType, encodedPayload string
	if r.Method != "GET" {
		h := sha256.Sum256(payload)
		encodedPayload = hex.EncodeToString(h[:])
		contentType = "content-type:" + r.Header.Get("Content-Type")
	}
	formattedBase := fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n%s",
		r.Method,
		strings.ToUpper(encodedPayload),
		contentType,
		fmt.Sprintf("date:%s", date.Format(http.TimeFormat)),
		ptiClientIDHeader+":"+r.Header.Get(ptiClientIDHeader),
		r.URL.Path,
	)

	// NB: These fields must be in the protected headers
	h := jws.NewHeaders()
	err := h.Set("cid", r.Header.Get(ptiClientIDHeader))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}
	err = h.Set("kid", publicKeyThumbprint)
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}
	signature, err := jws.Sign([]byte(formattedBase), jws.WithKey(jwa.RS512, key, jws.WithProtectedHeaders(h)))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	r.Header.Add(ptiSignatureHeader, string(signature))

	return nil
}

func VerifyBase(ctx context.Context, base []byte, r *http.Request) error {
	parts := strings.Split(string(base), "\n")
	if len(parts) < 6 {
		return fmt.Errorf("%w Signature base has incorrect format", ErrInvalidSignature)
	}

	if parts[0] != r.Method {
		return fmt.Errorf("%w Method mismatch", ErrInvalidSignature)
	}

	if r.Method != http.MethodGet {
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

		h := sha256.Sum256(body)

		if parts[1] != strings.ToUpper(hex.EncodeToString(h[:])) {
			return fmt.Errorf("%w Payload mismatch", ErrInvalidSignature)
		}
	}

	dateParts := strings.Split(parts[3], "date:")
	if len(dateParts) < 2 {
		return fmt.Errorf("%w Invalid date", ErrInvalidSignature)
	}
	date, err := time.Parse(http.TimeFormat, dateParts[1])
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	headerDate, err := time.Parse(http.TimeFormat, r.Header.Get("Date"))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}
	if !date.Equal(headerDate) {
		return fmt.Errorf("%w Signature date does not match date in the header.", ErrInvalidSignature)
	}

	if parts[5] != r.URL.Path {
		return fmt.Errorf("%w Path mismatch", ErrInvalidSignature)
	}

	return nil
}

func Verify(ctx context.Context, r *http.Request, key crypto.PublicKey) error {
	signature := r.Header.Get(ptiSignatureHeader)
	if signature == "" {
		return fmt.Errorf("%w Signature is empty", ErrInvalidSignature)
	}

	verifiedRawBase, err := jws.Verify([]byte(signature), jws.WithKey(jwa.RS512, key))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	return VerifyBase(ctx, verifiedRawBase, r)
}

func (c client) CreateUser(ctx context.Context, args CreateUserArgs) (string, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "pti"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "pti",
		})
	}

	url, err := url.JoinPath(c.baseURL, "users")
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)
	req.Header.Add("Content-Type", "application/json")
	date := time.Now()
	req.Header.Add("Date", date.Format(http.TimeFormat))
	if err := Sign(ctx, req, date, payload, c.privateKey, c.publicKeyThumbprint); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var userResp CreateUserResponse
	err = json.Unmarshal(body, &userResp)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return userResp.ID, nil
}

func (c client) CreateWallet(ctx context.Context, args CreateWalletArgs) (*Wallet, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "pti"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "pti",
		})
	}

	url, err := url.JoinPath(c.baseURL, "users", args.UserID, "wallets")
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)
	req.Header.Add("Content-Type", "application/json")
	date := time.Now()
	req.Header.Add("Date", date.Format(http.TimeFormat))
	if err := Sign(ctx, req, date, payload, c.privateKey, c.publicKeyThumbprint); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var wallet Wallet
	err = json.Unmarshal(body, &wallet)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &wallet, nil
}

func (c client) GetWallet(ctx context.Context, userID, id string) (*Wallet, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "GET"
		meta.Provider = "pti"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "GET",
			Provider: "pti",
		})
	}

	url, err := url.JoinPath(c.baseURL, "users", userID, "wallets", id)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)
	date := time.Now()
	req.Header.Add("Date", date.Format(http.TimeFormat))
	if err := Sign(ctx, req, date, nil, c.privateKey, c.publicKeyThumbprint); err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	var wallet Wallet
	err = json.Unmarshal(body, &wallet)
	if err != nil {
		return nil, fmt.Errorf("%w %s", ErrInternal, err)
	}

	return &wallet, nil
}

func (c client) StartUserAssessment(ctx context.Context, args CreateUserArgs) (string, error) {
	meta, ok := httplog.MetaForContext(ctx)
	if ok {
		meta.Method = "POST"
		meta.Provider = "pti"
	} else {
		ctx = context.WithValue(ctx, httplog.ContextKey, &httplog.Metadata{
			Method:   "POST",
			Provider: "pti",
		})
	}

	if args.ID == "" {
		return "", fmt.Errorf("%w UserID is required", ErrBadRequest)
	}

	url, err := url.JoinPath(c.baseURL, "users", "assessments")
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	payload, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	req.Header.Add(ptiClientIDHeader, c.clientID)
	req.Header.Add("Content-Type", "application/json")
	date := time.Now()
	req.Header.Add("Date", date.Format(http.TimeFormat))
	if err := Sign(ctx, req, date, payload, c.privateKey, c.publicKeyThumbprint); err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	resp, err := c.api.Do(req)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	err = checkResponseStatusCode(resp)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}
	defer resp.Body.Close()

	var userResp CreateUserResponse
	err = json.Unmarshal(body, &userResp)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	return userResp.ID, nil
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
