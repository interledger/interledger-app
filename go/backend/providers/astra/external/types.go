package external

import "time"

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

type CreateCardArgs struct {
	CardNumber       string `json:"card_number"`
	CardSecurityCode string `json:"card_security_code"`
	ExpirationDate   string `json:"expiration_date"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	StreetLine1      string `json:"street_line_1"`
	StreetLine2      string `json:"street_line_2"`
	City             string `json:"city"`
	State            string `json:"state"`
	ZipCode          string `json:"zip_code"`
	AddedByUser      bool   `json:"added_by_user"`
}

type CreateCardResp struct {
	ID              string    `json:"id"`
	AddressVerified bool      `json:"address_verified"`
	CardCompany     string    `json:"card_company"`
	City            string    `json:"city"`
	Created         time.Time `json:"created"`
	ExpirationDate  string    `json:"expiration_date"`
	FirstName       string    `json:"first_name"`
	FirstSixDigits  string    `json:"first_six_digits"`
	LastFourDigits  string    `json:"last_four_digits"`
	LastName        string    `json:"last_name"`
	PullEnabled     bool      `json:"pull_enabled"`
	PushEnabled     bool      `json:"push_enabled"`
	Removed         bool      `json:"removed"`
	ReviewStatus    string    `json:"review_status"`
	State           string    `json:"state"`
	Status          string    `json:"status"`
	StreetLine1     string    `json:"street_line_1"`
	StreetLine2     string    `json:"street_line_2"`
	ZipCode         string    `json:"zip_code"`
}

type AccountToCardArgs struct {
	Transfer        Transfer    `json:"transfer"`
	Source          Source      `json:"source"`
	Destination     Destination `json:"destination"`
	DebitFeePercent float64     `json:"debit_fee_percent"`
	Amount          string      `json:"amount"`
	Name            string      `json:"name"`
}
type Transfer struct {
	Type    string `json:"type"` // ach_credit |
	Addenda string `json:"addenda"`
}
type Source struct {
	ID string `json:"id"`
}
type Destination struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

type AccountToCardResp struct {
	ID              string      `json:"id"`
	Status          string      `json:"status"`
	Name            string      `json:"name"`
	AmountUSD       float64     `json:"amount"`
	Source          Source      `json:"source"`
	Destination     Destination `json:"destination"`
	StartDate       time.Time   `json:"start_date"`
	Created         time.Time   `json:"created"`
	Active          bool        `json:"active"`
	PaymentRoute    string      `json:"payment_route"`
	Type            string      `json:"type"`
	Blocked         bool        `json:"blocked"`
	DebitFeePercent float64     `json:"debit_fee_percent"`
}

type CardToAccountArgs struct {
	CardNumber       string `json:"card_number"`
	CardSecurityCode string `json:"card_security_code"`
	FirstName        string `json:"first_name"`
	ExpirationDate   string `json:"expiration_date"`
	LastName         string `json:"last_name"`
	StreetLine1      string `json:"street_line_1"`
	StreetLine2      string `json:"street_line_2"`
	City             string `json:"city"`
	State            string `json:"state"`
	ZipCode          string `json:"zip_code"`
	AddedByUser      bool   `json:"added_by_user"`
}

type CardToAccountResp struct {
	ID              string      `json:"id"`
	Status          string      `json:"status"`
	Name            string      `json:"name"`
	AmountUSD       float64     `json:"amount"`
	Source          Source      `json:"source"`
	Destination     Destination `json:"destination"`
	StartDate       time.Time   `json:"start_date"`
	Created         time.Time   `json:"created"`
	Active          bool        `json:"active"`
	PaymentRoute    string      `json:"payment_route"`
	Type            string      `json:"type"`
	Blocked         bool        `json:"blocked"`
	DebitFeePercent float64     `json:"debit_fee_percent"`
}
