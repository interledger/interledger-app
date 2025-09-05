package actions

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http/httptest"
	"os"

	"github.com/urfave/cli/v2"
	"gitlab.com/fynbos/httpmessagesignatures"
)

func MakeGenerateEd25519KeyPair(b Backends) cli.ActionFunc {
	return func(cCtx *cli.Context) error {
		pub, priv, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}

		privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
		if err != nil {
			return err
		}

		privPem := pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: privBytes,
		})

		err = os.WriteFile("private.pem", privPem, 0600)
		if err != nil {
			return err
		}

		pubJwk := map[string]string{
			"kty": "OKP",
			"crv": "Ed25519",
			"kid": "test",
			"x":   base64.StdEncoding.EncodeToString(pub),
		}

		marshalJWK, err := json.Marshal(pubJwk)
		if err != nil {
			return err
		}

		fmt.Println("Private key written to private.pem")
		fmt.Println("Public key (JWK)")
		fmt.Println(string(marshalJWK))

		return nil
	}
}

var SignGrantRequestFlags = []cli.Flag{
	&cli.StringFlag{
		Name:     "client",
		Usage:    "`client` url",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "resourceOwnerPaymentPointer",
		Aliases:  []string{"paymentPointer"},
		Usage:    "`client` url",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "privateKey",
		Usage:    "path to private key in PEM format.",
		Required: true,
	},
}

func MakeSignGrantRequest(b Backends) cli.ActionFunc {
	return func(cCtx *cli.Context) error {
		privPem, err := os.ReadFile(cCtx.String("privateKey"))
		if err != nil {
			return err
		}

		b, _ := pem.Decode([]byte(privPem))
		privKey, err := x509.ParsePKCS8PrivateKey(b.Bytes)
		if err != nil {
			return err
		}

		gr := map[string]interface{}{
			"client": cCtx.String("client"),
			"access_token": []map[string]interface{}{
				{
					"access": []map[string]interface{}{
						{
							"type":       "incoming-payment",
							"actions":    []string{"write", "read"},
							"identifier": cCtx.String("resourceOwnerPaymentPointer"),
						},
					},
					"label": "test",
				},
			},
		}
		body, err := json.Marshal(gr)
		if err != nil {
			return err
		}
		digest, err := httpmessagesignatures.CreateContentDigest(context.Background(), body, []string{"sha-256"})
		if err != nil {
			return err
		}

		err = httpmessagesignatures.VerifyContentDigest(context.Background(), digest, body)
		if err != nil {
			return err
		}

		req := httptest.NewRequest("POST", "http://auth.interledger.test/grant", bytes.NewBuffer(body))
		req.Header.Set("Content-Digest", digest)
		err = httpmessagesignatures.SignRequest(
			context.Background(),
			req,
			testEd25519Signer{privKey.(ed25519.PrivateKey)},
			[]string{"@method", "content-digest"},
			httpmessagesignatures.SignatureParams{
				Created: 1618884475,
				KeyID:   "test",
			},
			[]string{"content-digest"},
		)
		if err != nil {
			return err
		}

		fmt.Println("content-digest", req.Header.Get("content-digest"))
		fmt.Println("signature-input", req.Header.Get("signature-input"))
		fmt.Println("signature", req.Header.Get("signature"))
		fmt.Println("body", string(body))

		return nil
	}
}

type testEd25519Signer struct {
	privateKey ed25519.PrivateKey
}

func (s testEd25519Signer) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) ([]byte, error) {
	return ed25519.Sign(s.privateKey, digest), nil
}

func (s testEd25519Signer) Public() crypto.PublicKey {
	return nil
}
