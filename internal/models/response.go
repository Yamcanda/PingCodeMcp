package models

// Response represents the response structure from httpbin.org
type Response struct {
	Args    map[string]any    `json:"args"`
	Headers map[string]string `json:"headers"`
}

// ErrorResponse represents the error response structure from httpbin.org
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
