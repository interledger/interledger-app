package models

import "time"

// AccessToken represents an authentication token
type AccessToken struct {
	ID        string    `db:"id"`
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

// IsExpired checks if the token has expired
func (at *AccessToken) IsExpired() bool {
	return time.Now().After(at.ExpiresAt)
}

// IsValid checks if the token is valid and not expired
func (at *AccessToken) IsValid() bool {
	return at.Token != "" && !at.IsExpired()
}
