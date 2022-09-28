package kyc

import "time"

type Gender int

const (
	GenderUnknown Gender = 0
	GenderMale    Gender = 1
	GenderFemale  Gender = 2
)

type UserDetails struct {
	UserID      string `validate:"required,uuid"`
	FirstName   string
	LastName    string
	CountryCode string `validate:"omitempty,iso3166_1_alpha2"`
	Gender      Gender `validate:"omitempty,gte=0,lt=3"`
	DateOfBirth time.Time
	Address     *Address
}

type Address struct {
	Line1       string `json:"line_1,omitempty"`
	Line2       string `json:"line_2,omitempty"`
	Building    string `json:"building,omitempty"`
	Apartment   string `json:"apartment,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	ZipCode     string `json:"zip_code,omitempty"`
	CountryCode string `json:"country_code,omitempty" validate:"omitempty,iso3166_1_alpha2"`
}
