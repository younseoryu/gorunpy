package gorunpy

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
)

// Client provides a typed interface for calling Python functions.
// It manages the execution of a PyInstaller-built Python executable.
type Client struct {
	// binaryPath is the path to the Python executable.
	binaryPath string
}

// NewClient creates a new Client with the given binary path.
func NewClient(binaryPath string) *Client {
	return &Client{
		binaryPath: binaryPath,
	}
}

// BinaryPath returns the path to the Python executable.
func (c *Client) BinaryPath() string {
	return c.binaryPath
}

// Call executes a Python function with the given input and decodes the result.
// This is the low-level method used by generated typed methods.
//
// Parameters:
//   - ctx: Context for timeout and cancellation control.
//   - function: The name of the Python function to call.
//   - args: The function arguments as a map (will be JSON-encoded).
//   - result: A pointer to the result struct (will be JSON-decoded into).
//
// Returns an error if:
//   - The context is cancelled or times out
//   - JSON encoding/decoding fails
//   - The Python process fails
//   - The Python function returns an error
func (c *Client) Call(ctx context.Context, function string, args map[string]any, result any) error {
	// Build the request
	request := Request{
		Function: function,
		Args:     args,
	}

	// Encode the request to JSON
	requestJSON, err := json.Marshal(request)
	if err != nil {
		return &ErrJSONEncode{Err: err}
	}

	// Execute the Python process
	stdout, stderr, exitCode, err := c.execute(ctx, requestJSON)
	if err != nil {
		return err
	}

	// Handle the response based on exit code
	return c.handleResponse(function, stdout, stderr, exitCode, result)
}

// execute runs the Python executable with the given input.
func (c *Client) execute(ctx context.Context, input []byte) (stdout, stderr []byte, exitCode int, err error) {
	// Create the command with context
	cmd := exec.CommandContext(ctx, c.binaryPath)

	// Set up stdin
	cmd.Stdin = bytes.NewReader(input)

	// Capture stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Run the command
	runErr := cmd.Run()

	stdout = stdoutBuf.Bytes()
	stderr = stderrBuf.Bytes()

	// Check for context cancellation
	if ctx.Err() != nil {
		return nil, nil, -1, ctx.Err()
	}

	// Get the exit code
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Process failed to start or other error
			return nil, nil, -1, &ErrProcessFailed{
				Message:  runErr.Error(),
				ExitCode: -1,
				Stderr:   string(stderr),
			}
		}
	}

	return stdout, stderr, exitCode, nil
}

// handleResponse processes the Python response and populates the result.
func (c *Client) handleResponse(function string, stdout, stderr []byte, exitCode int, result any) error {
	switch exitCode {
	case ExitCodeSuccess:
		// Parse the success response
		var response Response
		if err := json.Unmarshal(stdout, &response); err != nil {
			return &ErrJSONDecode{Err: err, Output: string(stdout)}
		}

		if !response.OK {
			// This shouldn't happen with exit code 0, but handle it
			return &ErrProcessFailed{
				Message:  "response indicates failure but exit code was 0",
				ExitCode: exitCode,
				Stderr:   string(stderr),
			}
		}

		// Decode the result value
		if result != nil && response.Result != nil {
			// Re-encode the value and decode into the result type
			valueJSON, err := json.Marshal(response.Result.Value)
			if err != nil {
				return &ErrJSONDecode{Err: err, Output: string(stdout)}
			}
			if err := json.Unmarshal(valueJSON, result); err != nil {
				return &ErrJSONDecode{Err: err, Output: string(valueJSON)}
			}
		}

		return nil

	case ExitCodeHandledError, ExitCodeCrash:
		// Parse the error response from stderr
		var errResp ErrorResponse
		if err := json.Unmarshal(stderr, &errResp); err != nil {
			// Failed to parse error response
			return &ErrProcessFailed{
				Message:  "failed to parse error response",
				ExitCode: exitCode,
				Stderr:   string(stderr),
			}
		}

		if errResp.Error == nil {
			return &ErrProcessFailed{
				Message:  "error response missing error details",
				ExitCode: exitCode,
				Stderr:   string(stderr),
			}
		}

		// Map to appropriate Go error type
		pyErr := &PythonError{
			Kind:         ErrorKind(errResp.Error.Kind),
			Message:      errResp.Error.Message,
			Field:        errResp.Error.Field,
			FunctionName: function,
		}

		return mapPythonError(pyErr, exitCode)

	default:
		// Unknown exit code
		return &ErrProcessFailed{
			Message:  "unknown exit code",
			ExitCode: exitCode,
			Stderr:   string(stderr),
		}
	}
}

// CallRaw executes a Python function and returns the raw result value.
// This is useful when the result type is not known at compile time.
func (c *Client) CallRaw(ctx context.Context, function string, args map[string]any) (any, error) {
	var result any
	if err := c.Call(ctx, function, args, &result); err != nil {
		return nil, err
	}
	return result, nil
}

