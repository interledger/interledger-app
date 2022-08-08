package supporttickets

type CreateTicketArgs struct {
	FirstName   string `validate:"required"`
	LastName    string `validate:"required"`
	Email       string `validate:"required,email"`
	Description string `validate:"required"`
}
