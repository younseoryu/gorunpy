# GoRunPy

Call Python from Go with type safety.

```go
client := NewClient()
result, _ := client.Sum(ctx, 1, 2)  // calls Python, returns 3
```

## Quick Start

**1. Create project:**
```bash
# Create and enter project directory
mkdir myproject && cd myproject

# Initialize Go module
go mod init myproject
```

**2. Set up Python environment:**
```bash
# Create virtual environment
python3 -m venv venv

# Activate it
source venv/bin/activate

# Install gorunpy with build dependencies
pip install "gorunpy[build]"

# Add Go dependency
go get github.com/younseoryu/gorunpy/gorunpy
```

> **Note:** Keep the venv activated when running `go generate` — it needs the `gorunpy` CLI.

**3. Write Python:**
```bash
# Create Python package
mkdir mylib && touch mylib/__init__.py

# Create functions file with @gorunpy.export decorator
cat > mylib/functions.py << 'EOF'
import gorunpy

@gorunpy.export
def sum(a: int, b: int) -> int:
    return a + b
EOF
```

**4. Write Go:**
```bash
# Create main.go with go:generate directive
cat > main.go << 'EOF'
//go:generate gorunpy gen

package main

import (
	"context"
	"fmt"
)

func main() {
	client := NewClient()
	result, _ := client.Sum(context.Background(), 1, 2)
	fmt.Println(result) // 3
}
EOF
```

**5. Run:**
```bash
# Build Python binary + generate gorunpy_client.go
go generate

# Run the program
go run .
```

That's it! The `gorunpy gen` command:
- ✅ Auto-detects your Python module (searches up to 3 levels up/down)
- ✅ Auto-generates `__main__.py` if missing
- ✅ Builds the Python binary (hidden in `.gorunpy/`)
- ✅ Generates `gorunpy_client.go` with embedded binary
- ✅ Creates zero-config `NewClient()` function

## Advanced

### Specify module path
```bash
# Use specific module instead of auto-detect
gorunpy gen ./path/to/mylib
```

### Custom output locations
```bash
gorunpy gen --output ./bin --client ./pkg/client.go
```

### Without embedding (separate binary)
```bash
# Binary not embedded, must distribute separately
gorunpy gen --no-embed
```

### List functions in a binary
```bash
# Show available functions and signatures
gorunpy list .gorunpy/mylib
```

### Run a function directly
```bash
# Test without Go
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
