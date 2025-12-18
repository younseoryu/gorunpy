// Package gorunpy provides a Go-native, typed API for calling Python code
// that is packaged as a single executable (via PyInstaller).
//
// # Overview
//
// GoRunPy creates a typed executable boundary between Go and Python.
// It is NOT FFI, NOT embedding - it's process-based communication with
// full type safety on both sides.
//
// # Basic Usage
//
// Create a client and call Python functions:
//
//	client := gorunpy.NewClient("/path/to/python/executable")
//	
//	// Raw call with dynamic types
//	result, err := client.CallRaw(ctx, "sum", map[string]any{"a": 1, "b": 2})
//	
//	// Or use typed wrappers (see example subpackage)
//
// # Error Handling
//
// Errors are mapped to Go error types:
//
//   - [ErrInvalidInput]: Input validation or type errors
//   - [ErrUserCode]: Errors raised by Python user code
//   - [ErrRuntimeCrash]: Unhandled Python exceptions
//
// Context cancellation and timeouts are fully supported.
//
// # Creating Typed Clients
//
// For type-safe calls, create wrapper types and methods:
//
//	type SumInput struct {
//	    A int `json:"a"`
//	    B int `json:"b"`
//	}
//	
//	type SumOutput struct {
//	    Value int `json:"value"`
//	}
//	
//	func (c *MyClient) Sum(ctx context.Context, in SumInput) (SumOutput, error) {
//	    args := map[string]any{"a": in.A, "b": in.B}
//	    var value int
//	    if err := c.Call(ctx, "sum", args, &value); err != nil {
//	        return SumOutput{}, err
//	    }
//	    return SumOutput{Value: value}, nil
//	}
//
// # Python Side
//
// Install the gorunpy Python package and use the @export decorator:
//
//	from gorunpy import export
//	
//	@export
//	def sum(a: int, b: int) -> int:
//	    return a + b
//
// Then build with PyInstaller to create the executable.
//
// See https://github.com/younseoryu/gorunpy for full documentation.
package gorunpy

