# GoRunPy

Call Python from Go with type safety.

```go
client := NewClient()
result, _ := client.Sum(ctx, 1, 2)  // calls Python, returns 3
```

## Install

```bash
pip install gorunpy[build]
go get github.com/younseoryu/gorunpy
```

## Usage

**1. Write Python** (`mylib/functions.py`):
```python
import gorunpy

@gorunpy.export
def sum(a: int, b: int) -> int:
    return a + b
```

**2. Add to any Go file**:
```go
//go:generate gorunpy gen
```

**3. Build and run**:
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

## Example

```bash
cd example
python3 -m venv venv
source venv/bin/activate
pip install -e ../python[build]
go generate
go run .
```

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
your-project/
├── mylib/                    ← Your Python code
│   ├── __init__.py
│   └── functions.py          ← @gorunpy.export functions
├── main.go                   ← //go:generate gorunpy gen
├── gorunpy_client.go         ← Generated (embedded binary)
└── .gorunpy/                 ← Hidden build artifacts
    └── mylib                 ← Compiled binary
```

See [example/](./example/) for complete code.
