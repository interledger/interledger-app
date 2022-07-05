package unit

import (
	"encoding/json"
)

// This file should be used to represent the unit JSON api types
// https://docs.unit.co/

type (
	ResponseBody struct {
		Data []json.RawMessage `json:"data"`
	}
	RequestBody struct {
		Data json.RawMessage `json:"data"`
	}
	Tags struct {
		FynbosUserId string `json:"fynbosUserId,omitempty"`
	}

	// Webhooks

	EventType string
	Event     struct {
		ID   string    `json:"id"`
		Type EventType `json:"type"`
	}
	CustomerCreatedEvent struct {
		ID            string             `json:"id"`
		Type          string             `json:"type"`
		Attributes    EventAttributes    `json:"attributes"`
		Relationships EventRelationships `json:"relationships"`
	}

	ApplicationDeniedEvent struct {
		ID            string             `json:"id"`
		Type          string             `json:"type"`
		Attributes    EventAttributes    `json:"attributes"`
		Relationships EventRelationships `json:"relationships"`
	}

	EventAttributes struct {
		CreatedAt string `json:"createdAt"`
		Tags      Tags   `json:"tags"`
	}

	EventRelationships struct {
		Customer    JsonCustomer    `json:"customer,omitempty"`
		Application JsonApplication `json:"application"`
	}

	JsonCustomer struct {
		Data Data `json:"data"`
	}

	JsonApplication struct {
		Data Data `json:"data"`
	}

	Data struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}

	// Applications

	CreateApplicationRequest struct {
		Data CreateApplicationRequestData `json:"data"`
	}

	CreateApplicationRequestData struct {
		Type       string                       `json:"type"`
		Attributes RequestApplicationAttributes `json:"attributes"`
	}

	RequestApplicationAttributes struct {
		Ssn                string              `json:"ssn,omitempty"`
		Passport           string              `json:"passport,omitempty"`
		Nationality        string              `json:"nationality,omitempty"`
		FullName           FullName            `json:"fullName"`
		DateOfBirth        string              `json:"dateOfBirth"`
		Address            Address             `json:"address"`
		Phone              Phone               `json:"phone"`
		Email              string              `json:"email"`
		IP                 string              `json:"ip,omitempty"`
		Tags               Tags                `json:"tags,omitempty"`
		DeviceFingerprints []DeviceFingerprint `json:"deviceFingerprints,omitempty"`
	}

	CreateApplicationResponse struct {
		Data CreateApplicationResponseData `json:"data"`
	}
	CreateApplicationResponseData struct {
		ID            string                        `json:"id"`
		Type          string                        `json:"type"`
		Attributes    ResponseApplicationAttributes `json:"attributes"`
		Relationships ApplicationRelationships      `json:"relationships"`
	}

	ResponseApplicationAttributes struct {
		Ssn                string              `json:"ssn,omitempty"`
		Passport           string              `json:"passport,omitempty"`
		Nationality        string              `json:"nationality,omitempty"`
		FullName           FullName            `json:"fullName"`
		DateOfBirth        string              `json:"dateOfBirth"`
		Address            Address             `json:"address"`
		Phone              Phone               `json:"phone"`
		Email              string              `json:"email"`
		IP                 string              `json:"ip,omitempty"`
		SoleProprietorship bool                `json:"soleProprietorship,omitempty"`
		Ein                string              `json:"ein,omitempty"`
		Industry           string              `json:"industry,omitempty"`
		Dba                string              `json:"dba,omitempty"`
		Tags               Tags                `json:"tags,omitempty"`
		IdempotencyKey     string              `json:"idempotencyKey,omitempty"`
		DeviceFingerprints []DeviceFingerprint `json:"deviceFingerprints,omitempty"`
		Status             string              `json:"status,omitempty"`
		Archived           bool                `json:"archived,omitempty"`
	}

	FullName struct {
		First string `json:"first"`
		Last  string `json:"last"`
	}

	Address struct {
		Street     string `json:"street"`
		Street2    string `json:"street2,omitempty"`
		City       string `json:"city"`
		State      string `json:"state"`
		PostalCode string `json:"postalCode"`
		Country    string `json:"country"`
	}

	Phone struct {
		CountryCode string `json:"countryCode"`
		Number      string `json:"number"`
	}

	DeviceFingerprint struct {
		Provider string `json:"provider"`
		Value    string `json:"value"`
	}

	ApplicationRelationships struct {
		Customer    JsonCustomer    `json:"customer,omitempty"`
		Application JsonApplication `json:"application"`
	}
)
