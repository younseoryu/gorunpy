# GoRunPy

Call Python from Go with type safety.

```go
client := NewClient()
result, _ := client.Sum(ctx, 1, 2)  // calls Python, returns 3
```

## Quick Start

**1. Create project:**
```bash
mkdir myproject && cd myproject   # Create and enter project directory
go mod init myproject             # Initialize Go module
```

**2. Set up Python environment:**
```bash
python3 -m venv venv              # Create virtual environment
source venv/bin/activate          # Activate it
pip install "gorunpy[build]"      # Install gorunpy with build dependencies
go get github.com/younseoryu/gorunpy/gorunpy   # Add Go dependency
```

> **Note:** Keep the venv activated when running `go generate` — it needs the `gorunpy` CLI.

**3. Write Python:**
```bash
mkdir mylib && touch mylib/__init__.py   # Create Python package
```

```bash
cat > mylib/functions.py << 'EOF'        # Create functions file
import gorunpy

@gorunpy.export                          # Mark function as callable from Go
def sum(a: int, b: int) -> int:
    return a + b
EOF
```

**4. Write Go:**
```bash
cat > main.go << 'EOF'
//go:generate gorunpy gen                # Auto-detect, build, and generate client

package main

import (
	"context"
	"fmt"
)

func main() {
	client := NewClient()                                // Create client (uses embedded binary)
	result, _ := client.Sum(context.Background(), 1, 2)  // Call Python function
	fmt.Println(result)                                  // 3
}
EOF
```

**5. Run:**
```bash
go generate    # Builds Python binary + generates gorunpy_client.go
go run .       # Run the program
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
gorunpy gen ./path/to/mylib    # Use specific module instead of auto-detect
```

### Custom output locations
```bash
gorunpy gen --output ./bin --client ./pkg/client.go
```

### Without embedding (separate binary)
```bash
gorunpy gen --no-embed         # Binary not embedded, must distribute separately
```

### List functions in a binary
```bash
gorunpy list .gorunpy/mylib    # Show available functions and signatures
```

### Run a function directly
```bash
gorunpy run .gorunpy/mylib sum '{"a": 1, "b": 2}'   # Test without Go
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
