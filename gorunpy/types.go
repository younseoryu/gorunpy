package gorunpy

// Exit codes from Python executable.
const (
	exitSuccess = 0 // Success
	exitError   = 1 // Handled error (validation, user error)
	exitCrash   = 2 // Unhandled exception
)

// Request sent to Python.
type request struct {
	Function string         `json:"function"`
	Args     map[string]any `json:"args"`
}

// Response from Python (success).
type response struct {
	OK     bool   `json:"ok"`
	Result *value `json:"result,omitempty"`
}

type value struct {
	Value any `json:"value"`
}

// Response from Python (error).
type errorResponse struct {
	OK    bool         `json:"ok"`
	Error *errorDetail `json:"error,omitempty"`
}

type errorDetail struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}
