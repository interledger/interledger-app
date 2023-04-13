package persona

import (
	"encoding/json"
	"time"
)

type Webhook struct {
	Data WebhookData `json:"data"`
}

type Attributes struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
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
	CreatedAt             time.Time       `json:"created-at,omitempty"`
	UpdatedAt             time.Time       `json:"updated-at,omitempty"`
	RedactedAt            time.Time       `json:"redacted-at,omitempty"`
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
	Birthdate             time.Time       `json:"birthdate,omitempty"`
	SocialSecurityNumber  string          `json:"social-security-number,omitempty"`
	Tags                  []string        `json:"tags,omitempty"`
	IdentificationNumbers json.RawMessage `json:"identification-numbers,omitempty"`
}

type Inquiry struct {
	Data InquiryData `json:"data"`
}

type InquiryData struct {
	Type       string            `json:"type"`
	ID         string            `json:"id"`
	Attributes InquiryAttributes `json:"attributes"`
	Meta       InquiryMeta       `json:"meta"`
}

type InquiryAttributes struct {
	Status      string    `json:"status"`
	Subject     string    `json:"subject"`
	ReferenceID string    `json:"reference-id"`
	CreatedAt   time.Time `json:"created-at"`
	CompletedAt time.Time `json:"completed-at"`
	ExpiredAt   time.Time `json:"expired-at"`
}

type InquiryMeta struct {
	SessionToken string `json:"session-token"`
}

type AccountData struct {
	Data struct {
		Type       string               `json:"type"`
		ID         string               `json:"id"`
		Attributes IndividualAttributes `json:"attributes"`
	} `json:"data"`
}
