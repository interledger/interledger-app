package models

// APIResponse defines the standard Chimoney-like response envelope.
type APIResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	Data   any    `json:"data,omitempty"`
}
