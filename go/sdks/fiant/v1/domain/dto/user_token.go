package dto

// https://github.com/provenancetech/pti-platform-sdks/blob/master/java/src/main/java/com/pti/sdk/types/UserToken.java
type UserToken struct {
	AccessToken string  `json:"accessToken"`
	ExpiresAt   float64 `json:"expiresAt"`
	TokenType   string  `json:"tokenType"`
}
