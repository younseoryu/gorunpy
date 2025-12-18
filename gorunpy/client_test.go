package gorunpy_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/younseoryu/gorunpy/gorunpy"
)

func binaryPath(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("GORUNPY_TEST_BINARY"); p != "" {
		return p
	}
	wd, _ := os.Getwd()
	for _, p := range []string{
		filepath.Join(wd, "..", "example", "dist", "mylib"),
		filepath.Join(wd, "example", "dist", "mylib"),
	} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	t.Skip("test binary not found")
	return ""
}

func TestCall(t *testing.T) {
	c := gorunpy.NewClient(binaryPath(t))
	var result int
	err := c.Call(context.Background(), "sum", map[string]any{"a": 1, "b": 2}, &result)
	if err != nil {
		t.Fatal(err)
	}
	if result != 3 {
		t.Errorf("got %d, want 3", result)
	}
}

func TestErrValidation(t *testing.T) {
	c := gorunpy.NewClient(binaryPath(t))
	var result any
	err := c.Call(context.Background(), "sum", map[string]any{"a": "bad", "b": 2}, &result)
	var e *gorunpy.ErrValidation
	if !errors.As(err, &e) {
		t.Fatalf("got %T, want ErrValidation", err)
	}
	if e.Field != "a" {
		t.Errorf("field = %q, want %q", e.Field, "a")
	}
}

func TestErrNotFound(t *testing.T) {
	c := gorunpy.NewClient(binaryPath(t))
	var result any
	err := c.Call(context.Background(), "nonexistent", map[string]any{}, &result)
	var e *gorunpy.ErrNotFound
	if !errors.As(err, &e) {
		t.Fatalf("got %T, want ErrNotFound", err)
	}
}

func TestErrPython(t *testing.T) {
	c := gorunpy.NewClient(binaryPath(t))
	var result any
	err := c.Call(context.Background(), "divide", map[string]any{"a": 1.0, "b": 0.0}, &result)
	var e *gorunpy.ErrValidation
	if !errors.As(err, &e) {
		t.Fatalf("got %T, want ErrValidation", err)
	}
}

func TestContextCancel(t *testing.T) {
	c := gorunpy.NewClient(binaryPath(t))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var result any
	err := c.Call(ctx, "sum", map[string]any{"a": 1, "b": 2}, &result)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("got %v, want context.Canceled", err)
	}
}

func TestErrProcess(t *testing.T) {
	c := gorunpy.NewClient("/nonexistent")
	var result any
	err := c.Call(context.Background(), "sum", map[string]any{"a": 1, "b": 2}, &result)
	var e *gorunpy.ErrProcess
	if !errors.As(err, &e) {
		t.Fatalf("got %T, want ErrProcess", err)
	}
}
