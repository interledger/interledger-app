package external

const accessTokenID = "ad317668-0e30-4936-8b8d-b2517b2464fd"

var (
	IdentityTypeCorporate  = "corporate"
	IdentityTypeIndividual = "individual"
)

// todo: no longer used, remove
// type AccessToken struct {
// 	Token     string    `db:"token"`
// 	ExpiresAt time.Time `db:"expires_at"`
// }

// // todo: no longer used, remove
// func (ac *AccessToken) IsExpired() bool {
// 	return time.Now().After(ac.ExpiresAt)
// }
