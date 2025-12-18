package gorunpy

import "fmt"

type ErrorKind string

const (
	ErrorKindValidation       ErrorKind = "ValidationError"
	ErrorKindType             ErrorKind = "TypeError"
	ErrorKindFunctionNotFound ErrorKind = "FunctionNotFoundError"
	ErrorKindRuntime          ErrorKind = "RuntimeError"
)

type PythonError struct {
	Kind         ErrorKind `json:"kind"`
	Message      string    `json:"message"`
	Field        string    `json:"field,omitempty"`
	FunctionName string    `json:"-"`
}

func (e *PythonError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s, function: %s)", e.Kind, e.Message, e.Field, e.FunctionName)
	}
	return fmt.Sprintf("%s: %s (function: %s)", e.Kind, e.Message, e.FunctionName)
}

type ErrInvalidInput struct {
	*PythonError
}

func (e *ErrInvalidInput) Error() string {
	return fmt.Sprintf("invalid input: %s", e.PythonError.Error())
}

func (e *ErrInvalidInput) Unwrap() error {
	return e.PythonError
}

type ErrUserCode struct {
	*PythonError
}

func (e *ErrUserCode) Error() string {
	return fmt.Sprintf("user code error: %s", e.PythonError.Error())
}

func (e *ErrUserCode) Unwrap() error {
	return e.PythonError
}

type ErrRuntimeCrash struct {
	*PythonError
}

func (e *ErrRuntimeCrash) Error() string {
	return fmt.Sprintf("runtime crash: %s", e.PythonError.Error())
}

func (e *ErrRuntimeCrash) Unwrap() error {
	return e.PythonError
}

type ErrProcessFailed struct {
	Message  string
	ExitCode int
	Stderr   string
}

func (e *ErrProcessFailed) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("process failed (exit %d): %s: %s", e.ExitCode, e.Message, e.Stderr)
	}
	return fmt.Sprintf("process failed (exit %d): %s", e.ExitCode, e.Message)
}

type ErrJSONEncode struct {
	Err error
}

func (e *ErrJSONEncode) Error() string {
	return fmt.Sprintf("failed to encode input to JSON: %v", e.Err)
}

func (e *ErrJSONEncode) Unwrap() error {
	return e.Err
}

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

func mapPythonError(pyErr *PythonError, exitCode int) error {
	switch exitCode {
	case ExitCodeHandledError:
		switch pyErr.Kind {
		case ErrorKindValidation, ErrorKindType, ErrorKindFunctionNotFound:
			return &ErrInvalidInput{PythonError: pyErr}
		default:
			return &ErrUserCode{PythonError: pyErr}
		}
	case ExitCodeCrash:
		return &ErrRuntimeCrash{PythonError: pyErr}
	default:
		return &ErrRuntimeCrash{PythonError: pyErr}
	}
}
