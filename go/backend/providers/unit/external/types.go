package external

import (
	"encoding/json"
)

type (
	Response struct {
		Data   json.RawMessage `json:"data"`
		Errors []ResponseError `json:"errors"`
	}

	ResponseError struct {
		Title  string `json:"title"`
		Status string `json:"status"`
		Detail string `json:"detail"`
		Meta   struct {
			SupportId string `json:"supportId"`
		} `json:"meta"`
	}

	TypeData struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}

	Relationship struct {
		Data TypeData `json:"data"`
	}
)

type (
	ListApplicationFormRequest struct {
		Data []ApplicationForm `json:"data,omitempty"`
	}

	ApplicationFormRequest struct {
		Data ApplicationForm `json:"data,omitempty"`
	}

	ApplicationForm struct {
		ID         string                    `json:"id,omitempty"`
		Type       string                    `json:"type,omitempty"`
		Attributes ApplicationFormAttributes `json:"attributes,omitempty"`
	}

	ApplicationFormAttributes struct {
		Tags                    ApplicationTags        `json:"tags,omitempty"`
		ApplicationDetails      ApplicationFormPrefill `json:"applicationDetails,omitempty"`
		AllowedApplicationTypes []string               `json:"allowedApplicationTypes,omitempty"`
		Lang                    string                 `json:"lang,omitempty"`
		Url                     string                 `json:"url,omitempty"`
		Stage                   string                 `json:"stage,omitempty"`
	}

	ApplicationFormPrefill struct {
		ApplicationType    string              `json:"applicationType,omitempty"`
		Ssn                string              `json:"ssn,omitempty"`
		Passport           string              `json:"passport,omitempty"`
		Nationality        string              `json:"nationality,omitempty"`
		FullName           *FullName           `json:"fullName,omitempty"`
		DateOfBirth        string              `json:"dateOfBirth,omitempty"`
		Address            *Address            `json:"address,omitempty"`
		Phone              *Phone              `json:"phone,omitempty"`
		Email              string              `json:"email,omitempty"`
		IP                 string              `json:"ip,omitempty"`
		SoleProprietorship bool                `json:"soleProprietorship,omitempty"`
		Ein                string              `json:"ein,omitempty"`
		Industry           string              `json:"industry,omitempty"`
		Dba                string              `json:"dba,omitempty"`
		Tags               *ApplicationTags    `json:"tags,omitempty"`
		IdempotencyKey     string              `json:"idempotencyKey,omitempty"`
		DeviceFingerprints []DeviceFingerprint `json:"deviceFingerprints,omitempty"`
		Status             string              `json:"status,omitempty"`
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
)

type (
	ApplicationRequest struct {
		Data ApplicationWithoutRelationships `json:"data,omitempty"`
	}

	Application struct {
		ID            string                    `json:"id,omitempty"`
		Type          string                    `json:"type"`
		Attributes    ApplicationAttributes     `json:"attributes"`
		Relationships *ApplicationRelationships `json:"relationships,omitempty"`
	}

	ApplicationWithoutRelationships struct {
		ID         string                `json:"id,omitempty"`
		Type       string                `json:"type"`
		Attributes ApplicationAttributes `json:"attributes"`
	}

	ApplicationAttributes struct {
		Ssn                string              `json:"ssn"`
		Passport           string              `json:"passport,omitempty"`
		Nationality        string              `json:"nationality,omitempty"`
		FullName           *FullName           `json:"fullName,omitempty"`
		DateOfBirth        string              `json:"dateOfBirth"`
		Address            *Address            `json:"address,omitempty"`
		Phone              *Phone              `json:"phone,omitempty"`
		Email              string              `json:"email"`
		IP                 string              `json:"ip,omitempty"`
		Tags               *ApplicationTags    `json:"tags,omitempty"`
		DeviceFingerprints []DeviceFingerprint `json:"deviceFingerprints,omitempty"`
		IdempotencyKey     string              `json:"idempotencyKey,omitempty"`
		CreatedAt          string              `json:"createdAt,omitempty"`
		Status             string              `json:"status,omitempty"`
		Archived           bool                `json:"archived,omitempty"`
	}

	ApplicationRelationships struct {
		Customer        Relationship    `json:"customer,omitempty"`
		ApplicationForm ApplicationForm `json:"applicationForm,omitempty"`
	}
)

type (
	CounterpartyRequest struct {
		Data Counterparty `json:"data,omitempty"`
	}

	Counterparty struct {
		ID            string                    `json:"id,omitempty"`
		Type          string                    `json:"type"`
		Attributes    CounterpartyAttributes    `json:"attributes"`
		Relationships CounterpartyRelationships `json:"relationships"`
	}

	CounterpartyAttributes struct {
		Name           string `json:"name,omitempty"`
		RoutingNumber  string `json:"routingNumber,omitempty"`
		AccountNumber  string `json:"accountNumber,omitempty"`
		AccountType    string `json:"accountType,omitempty"`
		Type           string `json:"type,omitempty"`
		IdempotencyKey string `json:"idempotencyKey,omitempty"`
	}

	CounterpartyRelationships struct {
		Customer Relationship `json:"customer"`
	}
)

type (
	AchPayment struct {
		Type          string                  `json:"type,omitempty"`
		ID            string                  `json:"id,omitempty"`
		Attributes    AchPaymentAttributes    `json:"attributes,omitempty"`
		Relationships AchPaymentRelationships `json:"relationships,omitempty"`
	}

	AchPaymentAttributes struct {
		CreatedAt      string      `json:"createdAt,omitempty"`
		Status         string      `json:"status,omitempty"`
		Reason         interface{} `json:"reason,omitempty"`
		Description    string      `json:"description,omitempty"`
		Direction      string      `json:"direction,omitempty"`
		Amount         uint64      `json:"amount,omitempty"`
		Tags           DepositTags `json:"tags,omitempty"`
		IdempotencyKey string      `json:"idempotencyKey,omitempty"`
	}

	AchPaymentRelationships struct {
		Account      Relationship  `json:"account,omitempty"`
		Customer     *Relationship `json:"customer,omitempty"`
		Counterparty Relationship  `json:"counterparty,omitempty"`
	}

	AchPaymentRequest struct {
		Data AchPayment `json:"data"`
	}
)

const (
	AchStatusPending       = "Pending"
	AchStatusPendingReview = "PendingReview"
	AchStatusRejected      = "Rejected"
	AchStatusClearing      = "Clearing"
	AchStatusSent          = "Sent"
	AchStatusCanceled      = "Canceled"
	AchStatusReturned      = "Returned"
)

type (
	CreateDepositAccountRequest struct {
		Data DepositAccount `json:"data"`
	}

	DepositAccount struct {
		ID            string                       `json:"id,omitempty"`
		Type          string                       `json:"type,omitempty"`
		Attributes    DepositAccountAttributes     `json:"attributes,omitempty"`
		Relationships *DepositAccountRelationships `json:"relationships,omitempty"`
	}

	DepositAccountAttributes struct {
		CreatedAt        string `json:"createdAt,omitempty"`
		UpdatedAt        string `json:"updatedAt,omitempty"`
		Name             string `json:"name,omitempty"`
		DepositProduct   string `json:"depositProduct,omitempty"`
		RoutingNumber    string `json:"routingNumber,omitempty"`
		AccountNumber    string `json:"accountNumber,omitempty"`
		Currency         string `json:"currency,omitempty"`
		BalanceInCents   int64  `json:"balance,omitempty"`
		HoldInCents      int64  `json:"hold,omitempty"`
		AvailableInCents int64  `json:"available,omitempty"`
		Status           string `json:"status,omitempty"`
		Freezereason     string `json:"freezeReason,omitempty"`
		Closereason      string `json:"closeReason,omitempty"`
		FraudReason      string `json:"fraudReason,omitempty"`
		DacaStatus       string `json:"dacaStatus,omitempty"`
		IdempotencyKey   string `json:"idempotencyKey,omitempty"`
	}

	DepositAccountRelationships struct {
		Customer Relationship `json:"customer"`
	}
)

type (
	Statement struct {
		ID            string                 `json:"id,omitempty"`
		Type          string                 `json:"type,omitempty"`
		Attributes    StatementAttributes    `json:"attributes,omitempty"`
		Relationships StatementRelationships `json:"relationships,omitempty"`
	}

	StatementAttributes struct {
		Period string `json:"period,omitempty"`
	}

	StatementRelationships struct {
		Account   Relationship `json:"account"`
		Customer  Relationship `json:"customer"`
		Customers Relationship `json:"customers"`
	}
)
