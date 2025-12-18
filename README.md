# GoRunPy

Embed Python’s ML, AI, and data-science ecosystem directly into your Go binary with type-safe bindings. 
- No CGO
- No shared libraries
- No Python runtime

```python
@gorunpy.export
def add(a: int, b: int) -> int:
    return a + b
```

```go
client := NewPylibClient()
result, _ := client.Add(ctx, 1, 2)
fmt.Println(result) // 3
```

GoRunPy is ideal when you want to have Go as your primary system language but you want to call into Python for AI/ML or data workloads without operating a separate Python service. It’s designed for workloads where Python does meaningful work and sub-second call latency is acceptable.

## Quick Start

```bash
mkdir myproject && cd myproject
go mod init myproject
go run github.com/younseoryu/gorunpy/cmd/gorunpy@latest init
```

This creates:
- `.gorunpy/venv/` — Python environment with `gorunpy[build]`
- `pylib/calc.py` — Example Python functions
- `main.go` — Go code with `//go:generate`

Run it:

```bash
go generate && go run .
```

Output:
```
2 + 3 = 5
2.5 * 4.0 = 10.0
```

## Making Changes

Edit `pylib/calc.py` to add a new function:

```python
@gorunpy.export
def subtract(a: int, b: int) -> int:
    return a - b
```

Regenerate the client:

```bash
go generate
```

Use it in `main.go`:

```go
diff, _ := client.Subtract(ctx, 10, 3)
fmt.Printf("10 - 3 = %d\n", diff)
```

Run:

```bash
go run .
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

## When to Use

GoRunPy is ideal when you need Python libraries but don't want to manage a separate service:

- **Single binary deployment** — Python compiled and embedded; no runtime dependencies
- **Long-running operations** — ML inference, document processing
- **Complex Python ecosystems** — PyTorch, TensorFlow, Hugging Face, pandas, etc.
- **Simpler ops** — No separate Python service to deploy or monitor

**Not ideal for:**
- High-frequency calls — Use gRPC with a persistent Python service
- Simple logic that could be rewritten in Go

## CLI Reference

```bash
# Generate client (auto-detects Python module)
.gorunpy/venv/bin/gorunpy gen

# Specify module path
.gorunpy/venv/bin/gorunpy gen ./mylib

# Custom output locations
.gorunpy/venv/bin/gorunpy gen --output ./bin --client ./client.go

# Without embedding (requires distributing binary separately)
.gorunpy/venv/bin/gorunpy gen --no-embed

# List functions in a compiled binary
.gorunpy/venv/bin/gorunpy list .gorunpy/pylib

# Test a function directly
.gorunpy/venv/bin/gorunpy run .gorunpy/pylib add '{"a": 1, "b": 2}'
```

## How It Works

1. **Detection** — Finds Python packages with `@gorunpy.export`
2. **Build** — Compiles to standalone binary via PyInstaller
3. **Introspection** — Extracts function signatures
4. **Generation** — Creates type-safe Go client with `//go:embed`

## License

MIT

## TODO
1. clean up code
2. make it more convenient (a bit more opinionated?)
3. Better comments
