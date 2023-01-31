package authorisation

type Client struct {
	ID  string `db:"id"`
	URL string `db:"url"`
}

type Grant struct {
	ID    string
	State GrantState
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

type (
	GrantRequest struct {
		AccessToken AccessTokenReq `json:"access_token"`
		Client      ClientReq      `json:"client"`
		Interact    Interact       `json:"interact"`
		Subject     Subject        `json:"subject"`
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
		Kty string `json:"kty"`
		E   string `json:"e"`
		Kid string `json:"kid"`
		Alg string `json:"alg"`
		N   string `json:"n"`
	}

	Key struct {
		Proof string `json:"proof"`
		Jwk   Jwk    `json:"jwk"`
	}

	ClientReq struct {
		Display Display `json:"display"`
		Key     Key     `json:"key" validate:"url"`
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
		AccessToken AccessToken `json:"access_token"`
	}

	Access struct {
		Type      string   `json:"type"`
		Actions   []string `json:"actions"`
		Locations []string `json:"locations"`
		Datatypes []string `json:"datatypes"`
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
