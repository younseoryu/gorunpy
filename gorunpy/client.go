package gorunpy

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
)

type Client struct {
	binaryPath string
}

func NewClient(binaryPath string) *Client {
	return &Client{binaryPath: binaryPath}
}

func (c *Client) BinaryPath() string {
	return c.binaryPath
}

func (c *Client) Call(ctx context.Context, function string, args map[string]any, result any) error {
	request := Request{Function: function, Args: args}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return &ErrJSONEncode{Err: err}
	}

	stdout, stderr, exitCode, err := c.execute(ctx, requestJSON)
	if err != nil {
		return err
	}

	return c.handleResponse(function, stdout, stderr, exitCode, result)
}

func (c *Client) execute(ctx context.Context, input []byte) (stdout, stderr []byte, exitCode int, err error) {
	cmd := exec.CommandContext(ctx, c.binaryPath)
	cmd.Stdin = bytes.NewReader(input)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	runErr := cmd.Run()

	stdout = stdoutBuf.Bytes()
	stderr = stderrBuf.Bytes()

	if ctx.Err() != nil {
		return nil, nil, -1, ctx.Err()
	}

	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, nil, -1, &ErrProcessFailed{
				Message:  runErr.Error(),
				ExitCode: -1,
				Stderr:   string(stderr),
			}
		}
	}

	return stdout, stderr, exitCode, nil
}

func (c *Client) handleResponse(function string, stdout, stderr []byte, exitCode int, result any) error {
	switch exitCode {
	case ExitCodeSuccess:
		var response Response
		if err := json.Unmarshal(stdout, &response); err != nil {
			return &ErrJSONDecode{Err: err, Output: string(stdout)}
		}

		if !response.OK {
			return &ErrProcessFailed{
				Message:  "response indicates failure but exit code was 0",
				ExitCode: exitCode,
				Stderr:   string(stderr),
			}
		}

		if result != nil && response.Result != nil {
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
		var errResp ErrorResponse
		if err := json.Unmarshal(stderr, &errResp); err != nil {
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

		pyErr := &PythonError{
			Kind:         ErrorKind(errResp.Error.Kind),
			Message:      errResp.Error.Message,
			Field:        errResp.Error.Field,
			FunctionName: function,
		}

		return mapPythonError(pyErr, exitCode)

	default:
		return &ErrProcessFailed{
			Message:  "unknown exit code",
			ExitCode: exitCode,
			Stderr:   string(stderr),
		}
	}
}

func (c *Client) CallRaw(ctx context.Context, function string, args map[string]any) (any, error) {
	var result any
	if err := c.Call(ctx, function, args, &result); err != nil {
		return nil, err
	}
	return result, nil
}
