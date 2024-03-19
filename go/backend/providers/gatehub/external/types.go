package external

type Product string

var (
	OnOffRamp  Product = "OnOffRamp"
	Onboarding Product = "Onboarding"
	Exchange   Product = "Exchange"
)

type (
	IssueTokenReqeust struct {
		Scope []string `json:"scope"`
	}

	IssueTokenResponse struct {
		Token     string `json:"token,omitempty"`
		ExpiresAt string `json:"expires,omitempty"`
	}
)
