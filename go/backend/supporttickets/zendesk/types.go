package zendesk

type CreateTicketReq struct {
	Ticket Ticket `json:"ticket"`
}

type Comment struct {
	Body string `json:"body"`
}

type Requester struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Ticket struct {
	Subject   string    `json:"subject"`
	Requester Requester `json:"requester"`
	Comment   Comment   `json:"comment"`
}
