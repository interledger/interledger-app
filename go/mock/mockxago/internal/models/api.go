package models

// LoginRequest represents the login request body
type LoginRequest struct {
	PolicyID string      `json:"policyId"`
	Fields   []FieldData `json:"fields"`
}

// FieldData represents a field in the login request
type FieldData struct {
	FieldName  string `json:"fieldName"`
	FieldValue string `json:"fieldValue"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	TokenValue string `json:"tokenValue"`
}

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
