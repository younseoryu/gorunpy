package gorunpy

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
)

// Client calls Python functions in a PyInstaller executable.
type Client struct {
	path string
}

// NewClient creates a client for the Python executable at path.
func NewClient(path string) *Client {
	return &Client{path: path}
}

// Call invokes a Python function and decodes the result.
//
// Example:
//
//	var sum int
//	err := client.Call(ctx, "add", map[string]any{"a": 1, "b": 2}, &sum)
func (c *Client) Call(ctx context.Context, function string, args map[string]any, result any) error {
	reqJSON, err := json.Marshal(request{Function: function, Args: args})
	if err != nil {
		return &ErrJSON{Op: "encode", Err: err}
	}

	stdout, stderr, exitCode, err := c.exec(ctx, reqJSON)
	if err != nil {
		return err
	}

	return c.handle(stdout, stderr, exitCode, result)
}

// CallRaw invokes a Python function and returns the raw result.
func (c *Client) CallRaw(ctx context.Context, function string, args map[string]any) (any, error) {
	var result any
	err := c.Call(ctx, function, args, &result)
	return result, err
}

func (c *Client) exec(ctx context.Context, input []byte) ([]byte, []byte, int, error) {
	cmd := exec.CommandContext(ctx, c.path)
	cmd.Stdin = bytes.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	if ctx.Err() != nil {
		return nil, nil, -1, ctx.Err()
	}

	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return stdout.Bytes(), stderr.Bytes(), e.ExitCode(), nil
		}
		return nil, nil, -1, &ErrProcess{Message: err.Error(), ExitCode: -1, Stderr: stderr.String()}
	}

	return stdout.Bytes(), stderr.Bytes(), 0, nil
}

func (c *Client) handle(stdout, stderr []byte, exitCode int, result any) error {
	switch exitCode {
	case exitSuccess:
		var resp response
		if err := json.Unmarshal(stdout, &resp); err != nil {
			return &ErrJSON{Op: "decode", Err: err, Output: string(stdout)}
		}
		if result != nil && resp.Result != nil {
			b, _ := json.Marshal(resp.Result.Value)
			if err := json.Unmarshal(b, result); err != nil {
				return &ErrJSON{Op: "decode", Err: err, Output: string(b)}
			}
		}
		return nil

	case exitError, exitCrash:
		var resp errorResponse
		if err := json.Unmarshal(stderr, &resp); err != nil {
			return &ErrProcess{Message: "invalid error response", ExitCode: exitCode, Stderr: string(stderr)}
		}
		if resp.Error == nil {
			return &ErrProcess{Message: "missing error details", ExitCode: exitCode, Stderr: string(stderr)}
		}

		e := resp.Error
		switch e.Kind {
		case "ValidationError", "TypeError":
			return &ErrValidation{Message: e.Message, Field: e.Field}
		case "FunctionNotFoundError":
			return &ErrNotFound{Function: e.Message}
		default:
			return &ErrPython{Kind: e.Kind, Message: e.Message, Crash: exitCode == exitCrash}
		}

	default:
		return &ErrProcess{Message: "unknown exit code", ExitCode: exitCode, Stderr: string(stderr)}
	}
}
