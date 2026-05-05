package dto

import "time"

const accessTTL = 55 * time.Minute

type Access struct {
	Token     string    `db:"token"`
	ExpiresAt time.Time `db:"expires_at"`
}

func NewAccess(token string) Access {
	return Access{
		Token:     token,
		ExpiresAt: time.Now().Add(accessTTL),
	}
}

func (a Access) IsExpired() bool {
	return time.Now().After(a.ExpiresAt)
}

func (a Access) IsDifferent(other Access) bool {
	return a.Token != other.Token
}
