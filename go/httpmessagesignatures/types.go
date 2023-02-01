package httpmessagesignatures

import "crypto"

type SignatureParams struct {
	Created uint64
	Expires uint64
	Nonce   string
	Alg     string
	KeyID   string
	Tag     string
}

type Verifier interface {
	Verify(publicKey crypto.PublicKey, digest []byte, signature []byte) bool
}
