package ops

type (
	CreateAuthURLArgs struct {
		ClientID     string
		Scopes       []string
		RedirectURL  string
		AuthEndpoint string
	}
)
