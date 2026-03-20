package models

import "time"

// User represents a PTI user.
type User struct {
	ID                 string         `json:"id,omitempty"`
	Type               string         `json:"type,omitempty"`
	Status             string         `json:"status,omitempty"`
	StatusReason       string         `json:"statusReason,omitempty"`
	Tags               []string       `json:"tags,omitempty"`
	SourceOfFunds      string         `json:"sourceOfFunds,omitempty"`
	UserCreationDate   string         `json:"userCreateionDate,omitempty"`
	Addresses          []Address      `json:"addresses,omitempty"`
	UserPTIMetaData    map[string]any `json:"userPtiMeta,omitempty"`
	UserClientMetaData map[string]any `json:"userClientMeta,omitempty"`
	Emails             []Email        `json:"emails,omitempty"`
	Phones             []Phone        `json:"phones,omitempty"`
	Name               *Name          `json:"name,omitempty"`
	DateOfBirth        string         `json:"dateOfBirth,omitempty"`
	CreatedAt          time.Time      `json:"-"`
}

// Name represents a person's name.
type Name struct {
	First  string `json:"firstName,omitempty"`
	Last   string `json:"lastName,omitempty"`
	Middle string `json:"middleName,omitempty"`
}

// Email represents an email address.
type Email struct {
	Address string `json:"address,omitempty"`
	Default bool   `json:"default,omitempty"`
}

// Phone represents a phone number.
type Phone struct {
	Number  string `json:"number,omitempty"`
	Type    string `json:"type,omitempty"`
	Default bool   `json:"default,omitempty"`
}

// Address represents a physical address.
type Address struct {
	Street     string `json:"streetAddress,omitempty"`
	City       string `json:"city,omitempty"`
	PostalCode string `json:"postalCode,omitempty"`
	StateCode  string `json:"stateCode,omitempty"`
	Country    string `json:"country,omitempty"`
	Default    bool   `json:"default,omitempty"`
}

// Assessment represents a PTI user KYC assessment.
type Assessment struct {
	ResourceType  string `json:"resourceType"`
	ClientID      string `json:"clientId"`
	RequestID     string `json:"requestId"`
	UserID        string `json:"userId"`
	Date          string `json:"date"`
	Assessment    string `json:"assessment"`
	Tier          int    `json:"tier"`
	RefusalReason string `json:"refusalReason,omitempty"`
}

// TokenResponse represents a PTI JWT token response.
type TokenResponse struct {
	AccessToken string  `json:"accessToken"`
	ExpiresAt   float64 `json:"expiresAt"`
	TokenType   string  `json:"tokenType"`
}
