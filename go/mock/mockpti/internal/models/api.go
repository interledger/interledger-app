package models

// CreateUserRequest is the request body for POST /users.
type CreateUserRequest struct {
	ID                   string    `json:"id,omitempty"`
	Type                 string    `json:"type,omitempty"`
	DateOfBirth          string    `json:"dateOfBirth,omitempty"`
	Name                 Name      `json:"name,omitempty"`
	Emails               []Email   `json:"emails,omitempty"`
	Addresses            []Address `json:"addresses,omitempty"`
	Phones               []Phone   `json:"phones,omitempty"`
	SourceOfFunds        string    `json:"sourceOfFunds,omitempty"`
	CountryOfCitizenship string    `json:"countryOfCitizenship,omitempty"`
}

// PatchUserRequest is the request body for PATCH /users (merge).
type PatchUserRequest struct {
	ID            string    `json:"id,omitempty"`
	Type          string    `json:"type,omitempty"`
	DateOfBirth   string    `json:"dateOfBirth,omitempty"`
	Name          *Name     `json:"name,omitempty"`
	Emails        []Email   `json:"emails,omitempty"`
	Addresses     []Address `json:"addresses,omitempty"`
	Phones        []Phone   `json:"phones,omitempty"`
	SourceOfFunds string    `json:"sourceOfFunds,omitempty"`
}

// StartAssessmentRequest is the request body for POST /users/assessments.
type StartAssessmentRequest struct {
	ID            string    `json:"id,omitempty"`
	Type          string    `json:"type,omitempty"`
	DateOfBirth   string    `json:"dateOfBirth,omitempty"`
	Name          Name      `json:"name,omitempty"`
	Emails        []Email   `json:"emails,omitempty"`
	Addresses     []Address `json:"addresses,omitempty"`
	Phones        []Phone   `json:"phones,omitempty"`
	SourceOfFunds string    `json:"sourceOfFunds,omitempty"`
}

// TokenRequest is the request body for POST /auth/jwt.
type TokenRequest struct {
	URL    string `json:"url"`
	Method string `json:"method"`
}

// IDResponse is the common response for create operations.
type IDResponse struct {
	ID   string `json:"id"`
	Link string `json:"link,omitempty"`
}

// ErrorResponse is the standard error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
