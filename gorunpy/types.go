package gorunpy

// Exit codes from the Python executable.
const (
	// ExitCodeSuccess indicates successful execution.
	ExitCodeSuccess = 0
	// ExitCodeHandledError indicates a handled error (validation, user error).
	ExitCodeHandledError = 1
	// ExitCodeCrash indicates a crash or unhandled exception.
	ExitCodeCrash = 2
)

// Request represents a function call request to Python.
type Request struct {
	// Function is the name of the Python function to call.
	Function string `json:"function"`
	// Args contains the function arguments as key-value pairs.
	Args map[string]any `json:"args"`
}

// Response represents a successful response from Python.
type Response struct {
	// OK indicates whether the call succeeded.
	OK bool `json:"ok"`
	// Result contains the function return value on success.
	Result *ResultValue `json:"result,omitempty"`
}

// ResultValue wraps the actual return value.
type ResultValue struct {
	// Value is the actual return value from the Python function.
	Value any `json:"value"`
}

// ErrorResponse represents an error response from Python.
type ErrorResponse struct {
	// OK is always false for error responses.
	OK bool `json:"ok"`
	// Error contains the error details.
	Error *ErrorDetail `json:"error,omitempty"`
}

// ErrorDetail contains the details of an error.
type ErrorDetail struct {
	// Kind is the error category (e.g., "ValidationError", "TypeError").
	Kind string `json:"kind"`
	// Message is the error message.
	Message string `json:"message"`
	// Field is the field that caused the error (optional).
	Field string `json:"field,omitempty"`
}

