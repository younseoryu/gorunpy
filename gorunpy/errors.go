package gorunpy

import "fmt"

// Error types returned by Client methods.
//
// Use errors.As() to check error types:
//
//	var e *gorunpy.ErrValidation
//	if errors.As(err, &e) {
//	    fmt.Println(e.Field, e.Message)
//	}

// ErrValidation is returned when input validation fails.
// This includes type mismatches, missing arguments, and
// validation errors raised by Python code.
type ErrValidation struct {
	Message string // Error message
	Field   string // Which field failed (optional)
}

func (e *ErrValidation) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error: %s (field: %s)", e.Message, e.Field)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// ErrNotFound is returned when the requested function doesn't exist.
type ErrNotFound struct {
	Function string // Function name that was not found
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("function not found: %s", e.Function)
}

// ErrPython is returned when Python code raises an exception.
// This includes both handled errors and unhandled crashes.
type ErrPython struct {
	Kind    string // Exception type (e.g., "ValueError", "RuntimeError")
	Message string // Error message
	Crash   bool   // True if this was an unhandled exception
}

func (e *ErrPython) Error() string {
	if e.Crash {
		return fmt.Sprintf("python crash [%s]: %s", e.Kind, e.Message)
	}
	return fmt.Sprintf("python error [%s]: %s", e.Kind, e.Message)
}

// ErrProcess is returned when the Python process fails to execute.
// This includes binary not found, permission errors, timeouts, etc.
type ErrProcess struct {
	Message  string // What went wrong
	ExitCode int    // Process exit code (-1 if unknown)
	Stderr   string // Stderr output (if any)
}

func (e *ErrProcess) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("process error (exit %d): %s: %s", e.ExitCode, e.Message, e.Stderr)
	}
	return fmt.Sprintf("process error (exit %d): %s", e.ExitCode, e.Message)
}

// ErrJSON is returned when JSON encoding/decoding fails.
type ErrJSON struct {
	Op     string // "encode" or "decode"
	Err    error  // Underlying error
	Output string // Raw output (for decode errors)
}

func (e *ErrJSON) Error() string {
	if e.Output != "" {
		return fmt.Sprintf("json %s error: %v (output: %s)", e.Op, e.Err, e.Output)
	}
	return fmt.Sprintf("json %s error: %v", e.Op, e.Err)
}

func (e *ErrJSON) Unwrap() error {
	return e.Err
}
