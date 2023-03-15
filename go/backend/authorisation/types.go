package authorisation

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"time"
)

type Client struct {
	ID  string `db:"id"`
	URL string `db:"url"`
}

type Grant struct {
	ID            string
	Client        string
	State         GrantState
	Tokens        []AccessToken
	ContinueToken string
	Wait          string
}

type GrantState string

const (
	GrantStateProcessing GrantState = "processing"
	GrantStatePending    GrantState = "pending"
	GrantStateApproved   GrantState = "approved"
	GrantStateFinalized  GrantState = "finalized"
)

func (gs GrantState) IsValid() bool {
	return gs == GrantStateProcessing || gs == GrantStatePending || gs == GrantStateApproved || gs == GrantStateFinalized
}

func (gs GrantState) ValidTransition(next GrantState) bool {
	if !next.IsValid() {
		return false
	}

	switch gs {
	case GrantStateProcessing:
		return next == GrantStatePending || next == GrantStateFinalized || next == GrantStateApproved
	case GrantStatePending:
		return next == GrantStatePending || next == GrantStateFinalized || next == GrantStateProcessing
	case GrantStateApproved:
		return next == GrantStateProcessing || next == GrantStateFinalized || next == GrantStateApproved
	}

	return false
}

type TokenState string

const (
	TokenStateEnabled TokenState = "enabled"
	TokenStatePending TokenState = "pending"
)

type (
	GrantRequest struct {
		AccessToken []AccessTokenReq `json:"access_token"`
		Client      string           `json:"client"` // we identify a client by its payment pointer
		Interact    *Interact        `json:"interact,omitempty"`
		Subject     *Subject         `json:"subject,omitempty"`
	}

	AccessTokenReq struct {
		Access []Access `json:"access"`
		Label  string   `json:"label"`
	}

	Display struct {
		Name string `json:"name"`
		URI  string `json:"uri"`
	}

	Jwk struct {
		Kty       string    `json:"kty,omitempty"`
		E         string    `json:"e,omitempty"`
		Kid       string    `json:"kid,omitempty"`
		Alg       string    `json:"alg,omitempty"`
		N         string    `json:"n,omitempty"`
		Crv       string    `json:"crv,omitempty"`
		X         string    `json:"x,omitempty"`
		Use       string    `json:"use,omitempty"`
		ID        string    `json:"-"`
		CreatedAt time.Time `json:"-"`
		DeletedAt time.Time `json:"-"`
		ClientID  string    `json:"-"`
	}

	Key struct {
		Proof string `json:"proof"`
		Jwk   Jwk    `json:"jwk"`
	}

	Finish struct {
		Method string `json:"method"`
		URI    string `json:"uri"`
		Nonce  string `json:"nonce"`
	}

	Interact struct {
		Start  []string `json:"start"`
		Finish Finish   `json:"finish"`
	}

	Subject struct {
		SubIDFormats     []string `json:"sub_id_formats"`
		AssertionFormats []string `json:"assertion_formats"`
	}

	GrantAccessTokenResp struct {
		AccessTokens []AccessToken `json:"access_token"`
	}

	Access struct {
		Type       string   `json:"type"`
		Actions    []string `json:"actions"`
		Locations  []string `json:"locations,omitempty"`
		Datatypes  []string `json:"datatypes,omitempty"`
		Identifier string   `json:"identifier"`
	}

	AccessToken struct {
		Value     string   `json:"value"`
		Access    []Access `json:"access"`
		Label     string   `json:"label"`
		Manage    string   `json:"manage,omitempty"`
		ExpiresIn int      `json:"expires_in,omitempty"`
		Flags     []string `json:"flags,omitempty"`
	}
)

func (key Jwk) IsEdDSAPublicKey() bool {
	return key.Kty == "OKP" &&
		key.Crv == "Ed25519" &&
		key.X != ""
}

func (key Jwk) Fingerprint() (string, error) {
	if key.X == "" {
		return "", fmt.Errorf("%w Key is empty.", ErrInternal)
	}

	keyBytes, err := base64.StdEncoding.DecodeString(key.X)
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	pkixBytes, err := x509.MarshalPKIXPublicKey(ed25519.PublicKey(keyBytes))
	if err != nil {
		return "", fmt.Errorf("%w %s", ErrInternal, err)
	}

	hashedPem := sha256.Sum256(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pkixBytes,
	}))

	fingerprint := fmt.Sprintf("SHA256:%s", hex.EncodeToString(hashedPem[:]))

	return fingerprint, nil
}
