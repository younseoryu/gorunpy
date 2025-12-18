package gorunpy_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/younseoryu/gorunpy/gorunpy"
)

func testBinaryPath(t *testing.T) string {
	t.Helper()
	if path := os.Getenv("GORUNPY_TEST_BINARY"); path != "" {
		return path
	}
	wd, _ := os.Getwd()
	candidates := []string{
		filepath.Join(wd, "..", "example", "dist", "mathlib"),
		filepath.Join(wd, "example", "dist", "mathlib"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	t.Skip("test binary not found")
	return ""
}

func TestClientCallRaw(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))
	result, err := client.CallRaw(context.Background(), "sum", map[string]any{"a": 1, "b": 2})
	if err != nil {
		t.Fatalf("CallRaw failed: %v", err)
	}
	if val, ok := result.(float64); !ok || val != 3 {
		t.Errorf("expected 3, got %v", result)
	}
}

func TestClientCall(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))
	var result int
	err := client.Call(context.Background(), "sum", map[string]any{"a": 10, "b": 20}, &result)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if result != 30 {
		t.Errorf("expected 30, got %d", result)
	}
}

func TestTypeMismatchError(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))
	_, err := client.CallRaw(context.Background(), "sum", map[string]any{"a": "not an int", "b": 2})
	if err == nil {
		t.Fatal("expected error")
	}
	var e *gorunpy.ErrInvalidInput
	if !errors.As(err, &e) {
		t.Errorf("expected ErrInvalidInput, got %T", err)
	}
}

func TestFunctionNotFoundError(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))
	_, err := client.CallRaw(context.Background(), "nonexistent", map[string]any{})
	if err == nil {
		t.Fatal("expected error")
	}
	var e *gorunpy.ErrInvalidInput
	if !errors.As(err, &e) {
		t.Errorf("expected ErrInvalidInput, got %T", err)
	}
}

func TestValidationError(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))
	_, err := client.CallRaw(context.Background(), "divide", map[string]any{"a": 10.0, "b": 0.0})
	if err == nil {
		t.Fatal("expected error")
	}
	var e *gorunpy.ErrInvalidInput
	if !errors.As(err, &e) {
		t.Errorf("expected ErrInvalidInput, got %T", err)
	}
}

func TestContextCancellation(t *testing.T) {
	client := gorunpy.NewClient(testBinaryPath(t))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := client.CallRaw(ctx, "sum", map[string]any{"a": 1, "b": 2})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestInvalidBinaryPath(t *testing.T) {
	client := gorunpy.NewClient("/nonexistent")
	_, err := client.CallRaw(context.Background(), "sum", map[string]any{"a": 1, "b": 2})
	if err == nil {
		t.Fatal("expected error")
	}
	var e *gorunpy.ErrProcessFailed
	if !errors.As(err, &e) {
		t.Errorf("expected ErrProcessFailed, got %T", err)
	}
}
