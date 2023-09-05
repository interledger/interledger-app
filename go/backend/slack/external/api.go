package external

import (
	"context"

	"golang.org/x/oauth2"
)

type Client interface {
	GetConfig() *oauth2.Config
	CreateUserToken(ctx context.Context, authCode string) (*oauth2.Token, *User, error)
	GetAuthorizedUser(ctx context.Context, token *oauth2.Token) (*User, error)
}

type User struct {
	ID         string `json:"sub"`
	Username   string `json:"name"`
	TeamName   string `json:"https://slack.com/team_name"`
	TeamDomain string `json:"https://slack.com/team_domain"`
	TeamID     string `json:"https://slack.com/team_id"`
}
