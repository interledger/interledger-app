package v1

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jws"
)

func (art *apiRoundTripper) signature(req *http.Request) (string, error) {
	var contentType, encodedPayload string

	if req.Method != http.MethodGet {
		var payload []byte

		if req.Body != nil {
			var err error
			payload, err = io.ReadAll(req.Body)
			if err != nil {
				return "", err
			}
		}

		req.Body = io.NopCloser(bytes.NewReader((payload)))

		checksum := sha256.Sum256(payload)
		encodedPayload = hex.EncodeToString(checksum[:])
		contentType = "content-type:" + req.Header.Get("Content-Type")
	}

	// fmt.Printf("req method: %s\n", req.Method)

	sinatureTemplate := fmt.Sprintf(
		"%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		strings.ToUpper(encodedPayload),
		contentType,
		// fmt.Sprintf("date:%s", date.Format(http.TimeFormat)),
		fmt.Sprintf("date:%s", req.Header.Get("Date")),
		clientIDHeader+":"+req.Header.Get(clientIDHeader),
		req.URL.Path,
	)

	// NB: These fields must be in the protected headers
	h := jws.NewHeaders()

	if err := h.Set("cid", req.Header.Get(clientIDHeader)); err != nil {
		return "", fmt.Errorf("%w %s", ErrSettingCidInJWSHeaders, err)
	}
	if err := h.Set("kid", art.publicKeyThumbprint); err != nil {

		return "", fmt.Errorf("%w %s", ErrSettingKidInJWSHeaders, err)
	}
	signature, err := jws.Sign([]byte(sinatureTemplate), jws.WithKey(jwa.RS512(), art.privateKey, jws.WithProtectedHeaders(h)))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrComputingSignature, err)
	}

	return string(signature), nil
}
