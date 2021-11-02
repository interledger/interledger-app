package models

// Use the `json` tags to tell gqlgen what to use for the Graphql object.
type Organisation struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Created_at string
	Updated_at string
}
