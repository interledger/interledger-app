package kyc

import (
	"strings"
	"time"

	"github.com/interledger/interledger-app/go/backend/kyc/persona"
)

type Gender int

const (
	GenderUnknown Gender = 0
	GenderMale    Gender = 1
	GenderFemale  Gender = 2
	GenderOther   Gender = 3
)

type IndividualDetails struct {
	WalletID     string `db:"wallet_id" validate:"required,uuid"`
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	CountryCode  string `db:"country_code" validate:"omitempty,iso3166_1_alpha2"`
	PlaceOfBirth string
	Nationality  string `validate:"omitempty,iso3166_1_alpha2"`
	Gender       Gender `db:"gender" validate:"omitempty,gt=0,lt=4"`
	DateOfBirth  time.Time
	Address      *Address
	IPAddress    string `db:"ip_address" validate:"required,ip_addr"`
}

type Address struct {
	Line1       string `json:"line_1,omitempty"`
	Line2       string `json:"line_2,omitempty"`
	Building    string `json:"building,omitempty"`
	Apartment   string `json:"apartment,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty" validate:"omitempty,iso3166_2"`
	ZipCode     string `json:"zip_code,omitempty"`
	CountryCode string `json:"country_code,omitempty" validate:"omitempty,iso3166_1_alpha2"`
	PlaceID     string `json:"place_id,omitempty"`
}

//go:generate stringer -type=Status -trimprefix=Status

type Status int

const (
	StatusUnknown           Status = 0
	StatusPending           Status = 1
	StatusDocumentsRequired Status = 2
	StatusApproved          Status = 3
	StatusDenied            Status = 4
	StatusInReview          Status = 5
	StatusLevel1            Status = 6 // todo doc limits
	StatusLevel2            Status = 7 // todo doc limits
)

func (s Status) ToInt32() int32 {
	return int32(s)
}

func (a *Address) String() string {
	if a == nil {
		return ""
	}

	parts := []string{a.Line1, a.Line2, a.Apartment, a.Building, a.City, a.State, a.ZipCode, a.CountryCode}
	return strings.Join(parts, ",")
}

type PersonaInquiry struct {
	ID     string
	Status persona.InquiryStatus
}

type PersonaIDNumbers struct {
	SocialSecurity       string
	IssuingCountry       string
	IssuingState         string
	IdentificationClass  string
	IdentificationNumber string
	ExpirationDate       time.Time
}
