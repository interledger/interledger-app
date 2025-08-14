package persona

import (
	"context"
	"encoding/json"
	"time"

	"gitlab.com/fynbos/backend/country"
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
	ReferenceID           string                            `json:"reference-id,omitempty"`
	CreatedAt             string                            `json:"created-at,omitempty"`
	UpdatedAt             string                            `json:"updated-at,omitempty"`
	RedactedAt            string                            `json:"redacted-at,omitempty"`
	NameFirst             string                            `json:"name-first,omitempty"`
	NameMiddle            string                            `json:"name-middle,omitempty"`
	NameLast              string                            `json:"name-last,omitempty"`
	PhoneNumber           string                            `json:"phone-number,omitempty"`
	EmailAddress          string                            `json:"email-address,omitempty"`
	AddressStreet1        string                            `json:"address-street-1,omitempty"`
	AddressStreet2        string                            `json:"address-street-2,omitempty"`
	AddressCity           string                            `json:"address-city,omitempty"`
	AddressSubdivision    string                            `json:"address-subdivision,omitempty"`
	AddressPostalCode     string                            `json:"address-postal-code,omitempty"`
	CountryCode           string                            `json:"country-code,omitempty"`
	Birthdate             string                            `json:"birthdate,omitempty"`
	SocialSecurityNumber  string                            `json:"social-security-number,omitempty"`
	Tags                  []string                          `json:"tags,omitempty"`
	IdentificationNumbers map[string][]IdentificationNumber `json:"identification-numbers,omitempty"`
	InquiryTemplateID     string                            `json:"inquiry-template-id,omitempty"`
}

type IdentificationNumber struct {
	IssuingCountry       string    `json:"issuing-country"`
	IdentificationClass  string    `json:"identification-class"`
	IdentificationNumber string    `json:"identification-number"`
	CreatedAt            time.Time `json:"created-at"`
	UpdatedAt            time.Time `json:"updated-at"`
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
	AccountTagPending  AccountTag = "STATUS:PENDING"
	AccountTagReview   AccountTag = "STATUS:REVIEW"
	AccountTagVerified AccountTag = "STATUS:VERIFIED" // deprecated
	AccountTagRejected AccountTag = "STATUS:REJECTED"
	AccountTagLevel1   AccountTag = "STATUS:KYC-LEVEL:1"
	AccountTagLevel2   AccountTag = "STATUS:KYC-LEVEL:2"
)

type InquiryTemplateID string

var inquiryTemplateIDs = map[country.Country]InquiryTemplateID{
	country.US: "itmpl_JYHP6J5MtyaSd9UzcZRB2KBiPKCo", // only collect name, DOB and address
	country.GB: "itmpl_wAUZFdNhtuoQUoqf5uQmisEmX4qh",
	country.ZA: "itmpl_p9xFdzFCVrRSv7zbPUtbHLfSRFZQ",
}

func GetTemplateIDForCountry(ctx context.Context, ctry country.Country) InquiryTemplateID {
	id, ok := inquiryTemplateIDs[ctry]
	if !ok {
		return inquiryTemplateIDs[country.US] // fallback to US
	}

	return id
}

type ErrorResponse struct {
	Errors []Error `json:"errors"`
}

type Error struct {
	Title   string `json:"title"`
	Details string `json:"details"`
}
