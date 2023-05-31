package cmd

import (
	"crypto"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"gitlab.com/fynbos/cli/identities"
	"io"
	"os"
	"time"

	"github.com/spf13/viper"
)

type CreateOutgoingPaymentArgs struct {
	FromPP      string `json:"wallet"`
	Type        string `json:"type"`
	ToPP        string `json:"id"`
	ExternalRef string `json:"external_ref"`
	SendAmount  Amount `json:"send_amount"`
}

type OutgoingPayment struct {
	ID                string    `json:"id"`
	PaymentPointer    string    `json:"from"`
	ToPaymentPointer  string    `json:"to"`
	Failed            bool      `json:"failed"`
	Receiver          string    `json:"receiver"`
	SendAmount        Amount    `json:"send_amount"`
	ReceiveAmount     Amount    `json:"receive_amount"`
	SentAmount        Amount    `json:"sent_amount"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	FromLinkedAccount string    `json:"-"`
	CreatedBy         string    `json:"-"`
}

type IncomingPayment struct {
	ID                 string    `json:"id"`
	PaymentPointer     string    `json:"to"`
	FromPaymentPointer string    `json:"from"`
	IncomingAmount     *Amount   `json:"incoming_amount,omitempty"`
	ReceivedAmount     *Amount   `json:"outgoing_amount,omitempty"`
	Completed          bool      `json:"completed"`
	ExternalRef        string    `json:"external_ref"`
	Description        string    `json:"description"`
	ExpiresAt          time.Time `json:"expires_at"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	CreatedBy          string    `json:"-"`
}

type Amount struct {
	Amount   float64 `json:"amount,string"`
	Currency string  `json:"currency"`
}

type GrantState string

const (
	GrantStateProcessing GrantState = "processing"
	GrantStatePending    GrantState = "pending"
	GrantStateApproved   GrantState = "approved"
	GrantStateFinalized  GrantState = "finalized"
)

type (
	Grant struct {
		ID            string
		Client        string
		State         GrantState
		Tokens        []AccessToken
		ContinueToken string
		Wait          string
	}

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
		Kty string `json:"kty,omitempty"`
		E   string `json:"e,omitempty"`
		Kid string `json:"kid,omitempty"`
		Alg string `json:"alg,omitempty"`
		N   string `json:"n,omitempty"`
		Crv string `json:"crv,omitempty"`
		X   string `json:"x,omitempty"`
		Use string `json:"use,omitempty"`
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

	VerifyCommandArgs struct {
		Type          identities.Platform `validate:"required"`
		Identifier    string              `validate:"required"`
		WalletAddress string              `validate:"required,url"`
	}
)

type ed25519Signer struct {
	privateKey ed25519.PrivateKey
}

func (s ed25519Signer) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	return ed25519.Sign(s.privateKey, digest), nil
}

func (s ed25519Signer) Public() crypto.PublicKey {
	return s.privateKey.Public()
}

func NewEd25519Signer(config *viper.Viper) (*ed25519Signer, error) {
	privateKeyPem, err := os.ReadFile(config.GetString("clientKeyPath"))
	if err != nil {
		return nil, err
	}
	b, _ := pem.Decode(privateKeyPem)
	key, err := x509.ParsePKCS8PrivateKey(b.Bytes)
	if err != nil {
		return nil, err
	}

	privateKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("Failed to parse private key %s", config.GetString("clientKeyPath"))
	}

	return &ed25519Signer{privateKey}, nil
}
