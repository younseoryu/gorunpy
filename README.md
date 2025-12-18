# GoRunPy

Embed Python's ML, AI, and data science ecosystem into your Go binary with type-safe bindings.

```python
@gorunpy.export
def sum(a: int, b: int) -> int:
    return a + b
```

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

client := NewClient()
result, err := client.Sum(ctx, 1, 2)
if err != nil {
    log.Fatal(err)
}
fmt.Println(result) // 3
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
| `None` | (no return value) |

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

The `gorunpy gen` command handles the entire build pipeline:
- Auto-detects Python modules with `@gorunpy.export` (searches 3 levels up/down)
- Generates `__main__.py` entry point if missing
- Compiles Python to standalone binary via PyInstaller (output: `.gorunpy/`)
- Generates `gorunpy_client.go` with type-safe wrappers and embedded binary

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

## Generated Files

After running `go generate` (with venv activated), your project will contain:

```
myproject/
├── gorunpy_client.go         ← Generated client with embedded binary
└── .gorunpy/
    └── mylib                 ← Compiled Python binary
```

## When to Use

GoRunPy is ideal when you need Python libraries but don't want to manage a separate Python service:
- **Single binary deployment** — Python is compiled and embedded; no Python installation or sidecar service required
- **Long-running operations** — ML inference, document processing, data analysis where ~200-300ms call overhead is negligible
- **Complex Python ecosystems** — PyTorch, TensorFlow, Hugging Face, docling, pandas, etc.
- **Simpler ops** — No Python service to deploy, monitor, or scale separately

**Not ideal for:**
- High-frequency, low-latency calls — Use gRPC with a persistent Python service instead
- Simple operations that could be rewritten in Go

## License

MIT
