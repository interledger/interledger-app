package smileid

type Product string

const (
	EnhancedKYCProduct Product = "enhanced_kyc"
)

type (
	GetTokenResponse struct {
		Token string `json:"token"`
	}
)
