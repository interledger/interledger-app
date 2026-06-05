package signup

type Signup struct {
	ID           string
	UserID       string `db:"user_id"`
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	CountryCode  string `db:"country_code"`
	Email        string `db:"email"`
	MobileNumber string `db:"mobile_number"`
	Completed    bool
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
}

type MobileNumberArgs struct {
	ID           string `validate:"required,uuid"`
	MobileNumber string `validate:"required,e164"`
	OTP          string `validate:"required,numeric,len=6"`
}

type UserDataArgs struct {
	ID           string `validate:"omitempty,uuid"`
	FirstName    string `validate:"required"`
	LastName     string `validate:"required"`
	Email        string `validate:"required,email"`
	CountryCode  string `validate:"required,iso3166_1_alpha2"`
	MobileNumber string `validate:"required,e164"`
}
