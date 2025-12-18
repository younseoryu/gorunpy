package gorunpy_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/younseoryu/gorunpy/gorunpy"
)

// testBinaryPath returns the path to the test Python executable.
func testBinaryPath(t *testing.T) string {
	t.Helper()

	if path := os.Getenv("GORUNPY_TEST_BINARY"); path != "" {
		return path
	}

	// Try to find the example binary
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Go up from gorunpy/ to root, then into example/dist/
	candidates := []string{
		filepath.Join(wd, "..", "example", "dist", "mathlib"),
		filepath.Join(wd, "example", "dist", "mathlib"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	t.Skip("test binary not found, run 'cd example && make build-python' first or set GORUNPY_TEST_BINARY")
	return ""
}

func TestClientCallRaw(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))
	ctx := context.Background()

	result, err := client.CallRaw(ctx, "sum", map[string]any{"a": 1, "b": 2})
	if err != nil {
		t.Fatalf("CallRaw failed: %v", err)
	}

	// Result is float64 from JSON
	if val, ok := result.(float64); !ok || val != 3 {
		t.Errorf("expected 3, got %v", result)
	}
}

func TestClientCall(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))
	ctx := context.Background()

	var result int
	err := client.Call(ctx, "sum", map[string]any{"a": 10, "b": 20}, &result)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}

	if result != 30 {
		t.Errorf("expected 30, got %d", result)
	}
}

func TestTypeMismatchError(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))
	ctx := context.Background()

	_, err := client.CallRaw(ctx, "sum", map[string]any{
		"a": "not an int",
		"b": 2,
	})

	if err == nil {
		t.Fatal("expected error for type mismatch")
	}

	var invalidInput *gorunpy.ErrInvalidInput
	if !errors.As(err, &invalidInput) {
		t.Errorf("expected ErrInvalidInput, got %T: %v", err, err)
		return
	}

	if invalidInput.Kind != gorunpy.ErrorKindType {
		t.Errorf("expected TypeError, got %v", invalidInput.Kind)
	}

	if invalidInput.Field != "a" {
		t.Errorf("expected field 'a', got %v", invalidInput.Field)
	}
}

func TestFunctionNotFoundError(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))
	ctx := context.Background()

	_, err := client.CallRaw(ctx, "nonexistent_function", map[string]any{})

	if err == nil {
		t.Fatal("expected error for non-existent function")
	}

	var invalidInput *gorunpy.ErrInvalidInput
	if !errors.As(err, &invalidInput) {
		t.Errorf("expected ErrInvalidInput, got %T: %v", err, err)
	}
}

func TestValidationError(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))
	ctx := context.Background()

	// Division by zero should raise ValidationError
	_, err := client.CallRaw(ctx, "divide", map[string]any{"a": 10.0, "b": 0.0})

	if err == nil {
		t.Fatal("expected error for division by zero")
	}

	var invalidInput *gorunpy.ErrInvalidInput
	if !errors.As(err, &invalidInput) {
		t.Errorf("expected ErrInvalidInput, got %T: %v", err, err)
	}
}

func TestMissingArgument(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))
	ctx := context.Background()

	_, err := client.CallRaw(ctx, "sum", map[string]any{"a": 1})

	if err == nil {
		t.Fatal("expected error for missing argument")
	}

	var invalidInput *gorunpy.ErrInvalidInput
	if !errors.As(err, &invalidInput) {
		t.Errorf("expected ErrInvalidInput, got %T: %v", err, err)
	}
}

func TestContextCancellation(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.CallRaw(ctx, "sum", map[string]any{"a": 1, "b": 2})

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestInvalidBinaryPath(t *testing.T) {
	client := gorunpy.NewClient("/nonexistent/path/to/binary")
	ctx := context.Background()

	_, err := client.CallRaw(ctx, "sum", map[string]any{"a": 1, "b": 2})

	if err == nil {
		t.Fatal("expected error for invalid binary path")
	}

	var processErr *gorunpy.ErrProcessFailed
	if !errors.As(err, &processErr) {
		t.Errorf("expected ErrProcessFailed, got %T: %v", err, err)
	}
}

func TestBinaryPath(t *testing.T) {
	path := "/some/path"
	client := gorunpy.NewClient(path)

	if client.BinaryPath() != path {
		t.Errorf("BinaryPath() = %v, want %v", client.BinaryPath(), path)
	}
}

