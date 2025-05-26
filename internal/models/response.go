package models

// Response represents the response structure from httpbin.org
type Response struct {
	Args    map[string]any    `json:"args"`
	Headers map[string]string `json:"headers"`
}
