package external

import (
	"crypto"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jws"
)

func addSignatureHeader(
	r *http.Request,
	date time.Time,
	payload []byte,
	key crypto.PrivateKey,
	publicKeyThumbprint string,
) error {
	var contentType, encodedPayload string
	if r.Method != http.MethodGet {
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
	// signature, err := jws.Sign([]byte(formattedBase), jws.WithKey(jwa.RS512, key, jws.WithProtectedHeaders(h)))
	signature, err := jws.Sign([]byte(formattedBase), jws.WithKey(jwa.RS512(), key, jws.WithProtectedHeaders(h)))
	if err != nil {
		return fmt.Errorf("%w %s", ErrInternal, err)
	}

	r.Header.Add(ptiSignatureHeader, string(signature))

	return nil
}
