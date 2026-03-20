//go:build e2e
// +build e2e

package main

const (
	mockPTIURL      = "http://localhost:24025"
	maxWaitSeconds  = 60
	defaultClientID = "test-client-id"
)

// userResponse mirrors a subset of the PTI user JSON returned by the mock.
type userResponse struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

// assessmentResponse mirrors a subset of the PTI assessment JSON.
type assessmentResponse struct {
	RequestID  string `json:"requestId"`
	UserID     string `json:"userId"`
	Assessment string `json:"assessment"`
}

// walletResponse mirrors a subset of the PTI wallet JSON.
type walletResponse struct {
	WalletID string  `json:"walletId"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

// paymentInformationResponse mirrors a subset of the PTI payment-information JSON.
type paymentInformationResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// tokenResponse mirrors the PTI JWT response.
type tokenResponse struct {
	AccessToken string  `json:"accessToken"`
	TokenType   string  `json:"tokenType"`
	ExpiresAt   float64 `json:"expiresAt"`
}

// idResponse is the common create response.
type idResponse struct {
	ID   string `json:"id"`
	Link string `json:"link,omitempty"`
}

// errorResponse is the standard error body.
type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
