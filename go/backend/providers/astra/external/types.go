package external

type CreateIntentReq struct {
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Address1       string `json:"address1"`
	Address2       string `json:"address2"`
	City           string `json:"city"`
	State          string `json:"state"`
	PostalCode     string `json:"postal_code"`
	DateOfBirth    string `json:"date_of_birth"` // YYYY-MM-DD
	SocialSecurity string `json:"ssn"`
	IPAddress      string `json:"ip_address"`
}

type CreateIntentResp struct {
	ID string `json:"id"`
}

type Intent struct {
	ID                 string `json:"id"`
	UserID             string `json:"user_id"`
	Email              string `json:"email"`
	Phone              string `json:"phone"`
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	PreferredFirstName string `json:"preferred_first_name"`
	PreferredLastName  string `json:"preferred_last_name"`
	PreferredPronouns  string `json:"preferred_pronouns"`
	Status             string `json:"status"`
	KycType            string `json:"kyc_type"`
}

type GetVerificationTokenReq struct {
	Provider     string `json:"provider"`
	ProviderData struct {
		CustomerID string `json:"customer_id"`
	} `json:"provider_data"`
	ClientID     string `json:"client_id"`
	UserIntentID string `json:"user_intent_id"`
}

type VerificationTokenResp struct {
	Token string `json:"token"`
}

type AccessToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}
