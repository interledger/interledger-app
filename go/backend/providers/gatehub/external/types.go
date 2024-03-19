package external

import "time"

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

	CreateUserRequest struct {
		Email string `json:"email,omitempty"`
	}

	CreateUserResponse struct {
		ID                 string                 `json:"id"`
		CreatedAt          time.Time              `json:"createdAt"`
		UpdatedAt          time.Time              `json:"updatedAt"`
		ActivatedAt        time.Time              `json:"activatedAt"`
		Email              string                 `json:"email"`
		Secret2FA          bool                   `json:"secret2fa"`
		Type2FA            string                 `json:"type2fa"`
		Activated          bool                   `json:"activated"`
		Role               string                 `json:"role"`
		Meta               map[string]interface{} `json:"meta"`
		LastPasswordChange time.Time              `json:"lastPasswordChange"`
		Features           []string               `json:"features"`
		Managed            bool                   `json:"managed"`
		ManagedBy          string                 `json:"managedBy"`
	}
)
