package kyc

import "time"

type Gender int

const (
	GenderUnknown Gender = 0
	GenderMale    Gender = 1
	GenderFemale  Gender = 2
	GenderOther   Gender = 3
)

type UserDetails struct {
	UserID      string `db:"user_id" validate:"required,uuid"`
	FirstName   string `db:"first_name"`
	LastName    string `db:"last_name"`
	CountryCode string `db:"country_code" validate:"omitempty,iso3166_1_alpha2"`
	Gender      Gender `db:"gender" validate:"omitempty,gt=0,lt=4"`
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
