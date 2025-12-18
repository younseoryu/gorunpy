# GoRunPy

[![Go Reference](https://pkg.go.dev/badge/github.com/younseoryu/gorunpy.svg)](https://pkg.go.dev/github.com/younseoryu/gorunpy)

**Go-native, typed API for calling Python code packaged as a single executable.**

```go
// Calling Python feels exactly like calling Go
result, err := client.Sum(ctx, 1, 2)             // Returns 3
greeting, err := client.Greet(ctx, nil, "World") // Returns "Hello, World!"
```

## Installation

```bash
# Go client library
go get github.com/younseoryu/gorunpy

# Code generator (generates typed Go clients from Python)
go install github.com/younseoryu/gorunpy/cmd/gorunpy-gen@latest

# Python SDK
pip install gorunpy[build]
```

## How It Works

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                                                                              │
│   1. WRITE PYTHON              2. BUILD                3. GENERATE GO       │
│   ─────────────────            ─────────────────       ─────────────────    │
│                                                                              │
│   @export                      pyinstaller             gorunpy-gen          │
│   def sum(a: int, b: int)      --onefile               -binary ./mathlib    │
│       -> int:                  mathlib/                -package mathlib     │
│       return a + b             __main__.py             -output client.go    │
│                                                                              │
│                                     │                        │               │
│                                     ▼                        ▼               │
│                                                                              │
│                              dist/mathlib           mathlib/client.go       │
│                              (executable)           (generated)             │
│                                                                              │
│                                     │                        │               │
│                                     └──────────┬─────────────┘               │
│                                                │                             │
│                                                ▼                             │
│                                                                              │
│                                     4. USE IN GO                             │
│                                     ─────────────────                        │
│                                                                              │
│                                     client := mathlib.NewClient("./dist/mathlib")
│                                     result, _ := client.Sum(ctx, 1, 2)       │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

## Quick Start

See the complete working example in [`example/`](./example/README.md).

```bash
cd example
make setup      # Install Python dependencies
make build      # Build Python executable + generate Go client
make run        # Run the Go example
```

## Project Structure

```
gorunpy/
├── gorunpy/              # Go client library (go get this)
│   ├── client.go         # Client implementation
│   ├── errors.go         # Error types
│   └── types.go          # Protocol types
├── cmd/
│   └── gorunpy-gen/      # Code generator tool
├── python/
│   └── gorunpy/          # Python SDK (pip install this)
├── example/              # Complete working example
│   ├── python/           # Example Python code
│   ├── go/               # Example Go code (with generated client)
│   ├── Makefile          # Build commands
│   └── README.md         # Step-by-step guide
├── go.mod
└── README.md
```

## Usage

### Python Side

```python
# mathlib/functions.py
from gorunpy import export, ValidationError

@export
def sum(a: int, b: int) -> int:
    return a + b

@export
def divide(a: float, b: float) -> float:
    if b == 0:
        raise ValidationError("division by zero", field="b")
    return a / b
```

```python
# mathlib/__main__.py
import mathlib.functions
from gorunpy import main

if __name__ == "__main__":
    main()
```

### Build & Generate

```bash
# Build Python executable
pyinstaller --onefile --name mathlib mathlib/__main__.py

# Generate Go client (automatic!)
gorunpy-gen -binary dist/mathlib -package mathlib -output mathlib/client.go
```

### Go Side

```go
package main

import (
    "context"
    "fmt"
    "yourproject/mathlib"
)

func main() {
    client := mathlib.NewClient("./dist/mathlib")
    ctx := context.Background()

    // Type-safe calls - generated automatically!
    result, _ := client.Sum(ctx, 1, 2)
    fmt.Println(result) // 3
}
```

## Error Handling

```go
result, err := client.Divide(ctx, 10, 0)
if err != nil {
    var invalidInput *gorunpy.ErrInvalidInput
    if errors.As(err, &invalidInput) {
        fmt.Printf("Error on field %s: %s\n", invalidInput.Field, invalidInput.Message)
    }
}
```

## Supported Types

| Python | Go |
|--------|-----|
| `int` | `int` |
| `float` | `float64` |
| `str` | `string` |
| `bool` | `bool` |
| `List[T]` | `[]T` |
| `Dict[str, T]` | `map[string]T` |
| `Optional[T]` | `*T` |
| `Any` | `any` |

## Design Principles

> **This is NOT FFI. This is NOT embedding. This is a typed executable boundary.**

- ✅ No CGO
- ✅ No embedded Python
- ✅ No shared memory
- ✅ Stateless (per-call execution)
- ✅ Process isolation
- ✅ Full type safety

## License

MIT
