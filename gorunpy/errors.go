// Package gorunpy provides a Go-native, typed API for calling Python code
// that is packaged as a single executable (via PyInstaller).
package gorunpy

import (
	"fmt"
)

// ErrorKind represents the category of error from Python.
type ErrorKind string

const (
	// ErrorKindValidation indicates input validation failed.
	ErrorKindValidation ErrorKind = "ValidationError"
	// ErrorKindType indicates a type mismatch error.
	ErrorKindType ErrorKind = "TypeError"
	// ErrorKindFunctionNotFound indicates the requested function doesn't exist.
	ErrorKindFunctionNotFound ErrorKind = "FunctionNotFoundError"
	// ErrorKindRuntime indicates an unhandled runtime error.
	ErrorKindRuntime ErrorKind = "RuntimeError"
)

// PythonError represents an error returned from Python.
type PythonError struct {
	// Kind is the category of error.
	Kind ErrorKind `json:"kind"`
	// Message is the error message.
	Message string `json:"message"`
	// Field is the field that caused the error (for validation/type errors).
	Field string `json:"field,omitempty"`
	// FunctionName is the Python function that was being called.
	FunctionName string `json:"-"`
}

func (e *PythonError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s, function: %s)", e.Kind, e.Message, e.Field, e.FunctionName)
	}
	return fmt.Sprintf("%s: %s (function: %s)", e.Kind, e.Message, e.FunctionName)
}

// ErrInvalidInput indicates that input validation failed.
// This error type wraps validation and type errors from Python.
type ErrInvalidInput struct {
	*PythonError
}

func (e *ErrInvalidInput) Error() string {
	return fmt.Sprintf("invalid input: %s", e.PythonError.Error())
}

func (e *ErrInvalidInput) Unwrap() error {
	return e.PythonError
}

// ErrUserCode indicates that user code raised an exception.
// This includes validation errors raised intentionally by the Python function.
type ErrUserCode struct {
	*PythonError
}

func (e *ErrUserCode) Error() string {
	return fmt.Sprintf("user code error: %s", e.PythonError.Error())
}

func (e *ErrUserCode) Unwrap() error {
	return e.PythonError
}

// ErrRuntimeCrash indicates an unhandled exception or crash in Python.
type ErrRuntimeCrash struct {
	*PythonError
}

func (e *ErrRuntimeCrash) Error() string {
	return fmt.Sprintf("runtime crash: %s", e.PythonError.Error())
}

func (e *ErrRuntimeCrash) Unwrap() error {
	return e.PythonError
}

// ErrProcessFailed indicates that the Python process failed to execute.
type ErrProcessFailed struct {
	// Message describes what went wrong.
	Message string
	// ExitCode is the exit code of the process (-1 if not available).
	ExitCode int
	// Stderr contains any stderr output.
	Stderr string
}

func (e *ErrProcessFailed) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("process failed (exit %d): %s: %s", e.ExitCode, e.Message, e.Stderr)
	}
	return fmt.Sprintf("process failed (exit %d): %s", e.ExitCode, e.Message)
}

// ErrJSONEncode indicates that input could not be encoded to JSON.
type ErrJSONEncode struct {
	Err error
}

func (e *ErrJSONEncode) Error() string {
	return fmt.Sprintf("failed to encode input to JSON: %v", e.Err)
}

func (e *ErrJSONEncode) Unwrap() error {
	return e.Err
}

// ErrJSONDecode indicates that output could not be decoded from JSON.
type ErrJSONDecode struct {
	Err    error
	Output string
}

func (e *ErrJSONDecode) Error() string {
	return fmt.Sprintf("failed to decode output JSON: %v (output: %s)", e.Err, e.Output)
}

func (e *ErrJSONDecode) Unwrap() error {
	return e.Err
}

// mapPythonError maps a Python error to the appropriate Go error type.
func mapPythonError(pyErr *PythonError, exitCode int) error {
	switch exitCode {
	case ExitCodeHandledError:
		// Exit code 1: handled error (validation, user error)
		switch pyErr.Kind {
		case ErrorKindValidation, ErrorKindType:
			return &ErrInvalidInput{PythonError: pyErr}
		case ErrorKindFunctionNotFound:
			return &ErrInvalidInput{PythonError: pyErr}
		default:
			return &ErrUserCode{PythonError: pyErr}
		}
	case ExitCodeCrash:
		// Exit code 2: crash / unhandled exception
		return &ErrRuntimeCrash{PythonError: pyErr}
	default:
		// Unknown exit code, treat as crash
		return &ErrRuntimeCrash{PythonError: pyErr}
	}
}

