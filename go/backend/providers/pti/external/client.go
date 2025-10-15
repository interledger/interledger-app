package external

import (
	"bytes"
	"crypto"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
)

const (
	ptiProviderName = "pti"
)

type client struct {
	baseURL, clientID string

	// Thumbprint is used as the `kid` field in the jwt protected header
	publicKeyThumbprint string

	privateKey, publicKey jwk.Key

	http *http.Client
}

var _ Client = client{}

func sign(
	r *http.Request,
	date time.Time,
	payload []byte,
	key crypto.PrivateKey,
	publicKeyThumbprint string,
) error {
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
	signature, err := jws.Sign([]byte(formattedBase), jws.WithKey(jwa.RS512(), key, jws.WithProtectedHeaders(h)))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	r.Header.Add(ptiSignatureHeader, string(signature))

	return nil
}

// func VerifyBase(ctx context.Context, base []byte, r *http.Request) error {
// 	parts := strings.Split(string(base), "\n")
// 	if len(parts) < 6 {
// 		return fmt.Errorf("%w Signature base has incorrect format", ErrInvalidSignature)
// 	}

// 	if parts[0] != r.Method {
// 		return fmt.Errorf("%w Method mismatch", ErrInvalidSignature)
// 	}

// 	if r.Method != http.MethodGet {
// 		body, err := io.ReadAll(r.Body)
// 		if err != nil {
// 			return fmt.Errorf("%w %s", ErrInternal, err)
// 		}
// 		origRespBody := make([]byte, len(body))
// 		copy(origRespBody, body)
// 		defer func() {
// 			if body != nil {
// 				r.Body = io.NopCloser(bytes.NewBuffer(origRespBody))
// 			}
// 		}()

// 		h := sha256.Sum256(body)

// 		if parts[1] != strings.ToUpper(hex.EncodeToString(h[:])) {
// 			return fmt.Errorf("%w Payload mismatch", ErrInvalidSignature)
// 		}
// 	}

// 	dateParts := strings.Split(parts[3], "date:")
// 	if len(dateParts) < 2 {
// 		return fmt.Errorf("%w Invalid date", ErrInvalidSignature)
// 	}
// 	date, err := time.Parse(http.TimeFormat, dateParts[1])
// 	if err != nil {
// 		return fmt.Errorf("%w %s", ErrInternal, err)
// 	}

// 	headerDate, err := time.Parse(http.TimeFormat, r.Header.Get("Date"))
// 	if err != nil {
// 		return fmt.Errorf("%w %s", ErrInternal, err)
// 	}
// 	if !date.Equal(headerDate) {
// 		return fmt.Errorf("%w Signature date does not match date in the header", ErrInvalidSignature)
// 	}

// 	if parts[5] != r.URL.Path {
// 		return fmt.Errorf("%w Path mismatch", ErrInvalidSignature)
// 	}

// 	return nil
// }

// func Verify(ctx context.Context, r *http.Request, key crypto.PublicKey) error {
// 	signature := r.Header.Get(ptiSignatureHeader)
// 	if signature == "" {
// 		return fmt.Errorf("%w Signature is empty", ErrInvalidSignature)
// 	}

// 	verifiedRawBase, err := jws.Verify([]byte(signature), jws.WithKey(jwa.RS512, key))
// 	if err != nil {
// 		return fmt.Errorf("%w %s", ErrInternal, err)
// 	}

// 	return VerifyBase(ctx, verifiedRawBase, r)
// }

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
