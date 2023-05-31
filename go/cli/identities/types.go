package identities

type (
	JWKS struct {
		Keys []Jwk `json:"keys"`
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

	IdentityClaim struct {
		Wallet       string   `json:"wallet"`
		KeyID        string   `json:"kid"`
		Identifier   string   `json:"identifier"`
		Type         Platform `json:"type"`
		CreationTime string   `json:"ctime"`
	}

	VerifyArgs struct {
		Type          Platform `json:"type"`
		WalletAddress string   `json:"walletAddress"`
		Identifier    string   `json:"identifier"`
	}

	WalletDetails struct {
		ID         string     `json:"id"`
		PublicName string     `json:"publicName"`
		Identities []Identity `json:"platforms"`
	}

	Identity struct {
		Identifier    string   `json:"identifier"`
		KeyID         string   `json:"kid"`
		Type          Platform `json:"type"`
		CreationTime  string   `json:"ctime"`
		Signature     string   `json:"signature"`
		SignatureHash string   `json:"signature_hash"`
		PublicProof   string   `json:"public_proof"`
	}

	Platform string
)

const (
	PlatformTwitter Platform = "twitter"
)
