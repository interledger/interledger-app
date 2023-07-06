package persona

import (
	"encoding/json"
)

type Webhook struct {
	Data WebhookData `json:"data"`
}

type Attributes struct {
	Name      string          `json:"name"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt string          `json:"created-at"`
}

type WebhookData struct {
	Type       string     `json:"type"`
	ID         string     `json:"id"`
	Attributes Attributes `json:"attributes"`
}

type CreateInquiryReq struct {
	Data CreateInquiryReqData `json:"data"`
}

type CreateInquiryReqData struct {
	Attributes IndividualAttributes `json:"attributes"`
}

type IndividualAttributes struct {
	ReferenceID           string          `json:"reference-id,omitempty"`
	CreatedAt             string          `json:"created-at,omitempty"`
	UpdatedAt             string          `json:"updated-at,omitempty"`
	RedactedAt            string          `json:"redacted-at,omitempty"`
	NameFirst             string          `json:"name-first,omitempty"`
	NameMiddle            string          `json:"name-middle,omitempty"`
	NameLast              string          `json:"name-last,omitempty"`
	PhoneNumber           string          `json:"phone-number,omitempty"`
	EmailAddress          string          `json:"email-address,omitempty"`
	AddressStreet1        string          `json:"address-street-1,omitempty"`
	AddressStreet2        string          `json:"address-street-2,omitempty"`
	AddressCity           string          `json:"address-city,omitempty"`
	AddressSubdivision    string          `json:"address-subdivision,omitempty"`
	AddressPostalCode     string          `json:"address-postal-code,omitempty"`
	CountryCode           string          `json:"country-code,omitempty"`
	Birthdate             string          `json:"birthdate,omitempty"`
	SocialSecurityNumber  string          `json:"social-security-number,omitempty"`
	Tags                  []string        `json:"tags,omitempty"`
	IdentificationNumbers json.RawMessage `json:"identification-numbers,omitempty"`
	InquiryTemplateID     string          `json:"inquiry-template-id,omitempty"`
}

type Inquiry struct {
	Data     InquiryData       `json:"data"`
	Included []InquiryIncluded `json:"included"`
}

type InquiryData struct {
	Type       string            `json:"type"`
	ID         string            `json:"id"`
	Attributes InquiryAttributes `json:"attributes"`
	Meta       InquiryMeta       `json:"meta"`
}

type InquiryAttributes struct {
	Status               string `json:"status"`
	Subject              string `json:"subject"`
	ReferenceID          string `json:"reference-id"`
	CreatedAt            string `json:"created-at"`
	CompletedAt          string `json:"completed-at"`
	ExpiredAt            string `json:"expired-at"`
	SocialSecurityNumber string `json:"social-security-number"`
}

type InquiryIncluded struct {
	Type       string                    `json:"type"`
	Attributes InquiryIncludedAttributes `json:"attributes"`
}

type InquiryIncludedAttributes struct {
	IPAddress          string `json:"ip-address"`
	IDClass            string `json:"id-class"`
	IDNumber           string `json:"identification-number"`
	CountryCode        string `json:"country-code"`
	AddressSubdivision string `json:"address-subdivision"`
	ExpirationDate     string `json:"expiration-date"`
	Birthplace         string `json:"birthplace"`
	Nationality        string `json:"nationality"`
	Gender             string `json:"sex"`
}

type InquiryMeta struct {
	SessionToken string `json:"session-token"`
}

type InquiryStatus string

const (
	InquiryStarted     InquiryStatus = "started"
	InquiryCreated     InquiryStatus = "created"
	InquiryPending     InquiryStatus = "pending"
	InquiryExpired     InquiryStatus = "expired"
	InquiryCompleted   InquiryStatus = "completed"
	InquiryFailed      InquiryStatus = "failed"
	InquiryNeedsReview InquiryStatus = "needs_review"
	InquiryApproved    InquiryStatus = "approved"
	InquiryDeclined    InquiryStatus = "declined"
)

type Account struct {
	Data AccountData `json:"data"`
}

type AccountData struct {
	Type       string               `json:"type"`
	ID         string               `json:"id"`
	Attributes IndividualAttributes `json:"attributes"`
}

type AccountTag string

const (
	AccountTagDirty    AccountTag = "DIRTY"
	AccountTagPending  AccountTag = "PENDING"
	AccountTagReview   AccountTag = "REVIEW"
	AccountTagVerified AccountTag = "VERIFIED"
	AccountTagRejected AccountTag = "REJECTED"
)
