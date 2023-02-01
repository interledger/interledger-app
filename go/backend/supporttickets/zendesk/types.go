package zendesk

type CreateRequestReq struct {
	Request Request `json:"request"`
}

type Comment struct {
	Body string `json:"body"`
}

type Requester struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Request struct {
	Subject   string    `json:"subject"`
	Requester Requester `json:"requester"`
	Comment   Comment   `json:"comment"`
}
