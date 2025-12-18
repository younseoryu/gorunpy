# GoRunPy

Call Python from Go with type safety.

```go
client := NewClient()
result, _ := client.Sum(ctx, 1, 2)  // calls Python, returns 3
```

## Quick Start

**1. Create project:**
```bash
mkdir myproject && cd myproject
go mod init myproject
```

**2. Set up Python environment:**
```bash
python3 -m venv venv
source venv/bin/activate
pip install "gorunpy[build]"
go get github.com/younseoryu/gorunpy/gorunpy
```

> **Note:** Keep the venv activated when running `go generate` — it needs the `gorunpy` CLI.

**3. Write Python** (`mylib/functions.py`):
```bash
mkdir mylib && touch mylib/__init__.py
```

```python
# mylib/functions.py
import gorunpy

@gorunpy.export
def sum(a: int, b: int) -> int:
    return a + b
```

**4. Write Go** (`main.go`):
```go
//go:generate gorunpy gen

package main

import (
	"context"
	"fmt"
)

func main() {
	client := NewClient()
	result, _ := client.Sum(context.Background(), 1, 2)
	fmt.Println(result)  // 3
}
```

**5. Run:**
```bash
go generate
go run .
```

That's it! The `gorunpy gen` command:
- ✅ Auto-detects your Python module
- ✅ Auto-generates `__main__.py` if missing
- ✅ Builds the Python binary (hidden in `.gorunpy/`)
- ✅ Generates `gorunpy_client.go` with embedded binary
- ✅ Creates zero-config `NewClient()` function

## Advanced

### Specify module path
```bash
gorunpy gen ./path/to/mylib
```

### Custom output locations
```bash
gorunpy gen --output ./bin --client ./pkg/client.go
```

### Without embedding (separate binary)
```bash
gorunpy gen --no-embed
```

### List functions in a binary
```bash
gorunpy list .gorunpy/mylib
```

### Run a function directly
```bash
gorunpy run .gorunpy/mylib sum '{"a": 1, "b": 2}'
```

## How It Works

1. **Detection**: Scans for Python packages with `@gorunpy.export` decorated functions
2. **Build**: Uses PyInstaller to create a standalone binary
3. **Introspection**: Queries the binary for function signatures
4. **Generation**: Creates Go client code with type-safe wrappers
5. **Embedding**: Binary is embedded via `//go:embed` for single-binary distribution

## Project Structure

```
myproject/
├── venv/                     ← Python virtual environment
├── mylib/                    ← Your Python code
│   ├── __init__.py
│   └── functions.py          ← @gorunpy.export functions
├── go.mod
├── main.go                   ← //go:generate gorunpy gen
├── gorunpy_client.go         ← Generated (embedded binary)
└── .gorunpy/                 ← Hidden build artifacts
    └── mylib                 ← Compiled binary
```

See [example/](./example/) for complete code.
