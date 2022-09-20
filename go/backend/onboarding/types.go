package onboarding

type GetOnboardingArgs struct {
	Id string `validate:"required,uuid"`
}

type UpdateOnboardingArgs struct {
	Id               string `validate:"omitempty,uuid"`
	Country          string `validate:"omitempty,iso3166_1_alpha2"`
	FirstName        string `validate:"omitempty"`
	LastName         string `validate:"omitempty"`
	Email            string `validate:"omitempty,email"`
	Phone            string `validate:"omitempty,e164"`
	PhoneVerified    bool   `validate:"omitempty"`
	ServiceAgreement bool   `validate:"omitempty"`
}

type CreateAccountArgs struct {
	IdentityID string `validate:"required,uuid"`
	Country    string `validate:"required,iso3166_1_alpha2"`
}

type Onboarding struct {
	ID               string
	FirstName        string `db:"first_name"`
	LastName         string `db:"last_name"`
	Country          string `db:"country_of_residence"`
	Email            string `db:"email"`
	Phone            string `db:"phone"`
	PhoneVerified    bool   `db:"phone_verified"`
	ServiceAgreement bool   `db:"service_agreement"`
	CreatedAt        string `db:"created_at"`
	UpdatedAt        string `db:"updated_at"`
}
