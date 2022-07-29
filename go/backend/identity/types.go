package identity

import "fmt"

type Identity struct {
	ID           string
	FirstName    string `db:"first_name"`
	LastName     string `db:"last_name"`
	MobileNumber string `db:"mobile_number"`
	Email        string
	DateOfBirth  string
	Country      string
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
}

type CreateArgs struct {
	ID           string `validate:"required,uuid"`
	FirstName    string `validate:"required"`
	LastName     string `validate:"required"`
	MobileNumber string `validate:"required"` // TODO: decide on format
	Email        string `validate:"required,email"`
	Country      string `validate:"required,iso3166_1_alpha2"`
}

// We can control what information is allowed to be logged from here.
// TODO: decide what is sensitive information. Might be better to implement at Zap level?
func (args CreateArgs) String() string {
	return fmt.Sprintf("id=%s,firstName=%s,lastName=%s,mobileNum=%s,email=%s,country=%s",
		args.ID,
		args.FirstName,
		args.LastName,
		args.MobileNumber,
		args.Email,
		args.Country,
	)
}
