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

## Examples

| Example | Library | Use Case |
|---------|---------|----------|
| [docling](examples/docling/) | Docling | PDF/document to Markdown conversion |
| [pandas](examples/pandas/) | pandas | CSV analysis, filtering, aggregation |
| [sentiment](examples/sentiment/) | Transformers | Text sentiment analysis with HuggingFace |
| [image](examples/image/) | Pillow | Image resize, convert, filter |

```bash
# Try an example
cd examples/pandas
python -m venv .venv && source .venv/bin/activate  # or python3 -m venv .venv
pip install "gorunpy[build]" pandas jinja2
gorunpy
go run .
```

## CLI Reference

### Initialize a new project

```bash
go run github.com/younseoryu/gorunpy/cmd/gorunpy@latest init
```

### Regenerate after changes

```bash
go generate
```

### Customizing generation

Edit the `//go:generate` line in your `main.go`:

**By default, this:**
- Reads `pylib/`
- Writes binary to `.gorunpy/pylib`
- Writes `gorunpy_client.go`

```go
//go:generate .gorunpy/venv/bin/gorunpy
```

**Common customizations:**

Read from `./mymodule` instead of `pylib/`:

```go
//go:generate .gorunpy/venv/bin/gorunpy ./mymodule
```

Write binary to `./bin` instead of `.gorunpy/` (Go client automatically embeds from there):

```go
//go:generate .gorunpy/venv/bin/gorunpy -o ./bin
```

Write Go client to `./pkg/client.go` instead of `gorunpy_client.go`:

```go
//go:generate .gorunpy/venv/bin/gorunpy --client ./pkg/client.go
```

## How It Works

1. **Detection** — Finds Python packages with `@gorunpy.export`
2. **Build** — Compiles to standalone binary via PyInstaller
3. **Introspection** — Extracts function signatures
4. **Generation** — Creates type-safe Go client with `//go:embed`

## Comparison
|                     | GoRunPy                  | gRPC / HTTP           | CGO bindings        |
|---------------------|--------------------------|-----------------------|---------------------|
| Deployment          | Single binary            | Separate services     | Single binary       |
| CGO Required        | ❌ No                     | ❌ No                  | ✅ Yes              |
| Python Runtime      | Bundled (out-of-process) | Required (service)    | Embedded (in-proc)  |
| Cross-compilation   | ⚠️ Limited (Python binary must be built per OS)                  | ✅ Easy                | ❌ Hard             |
| Typical Latency     | ~200–500 ms              | ~1–10 ms              | ~µs-level           |
| Throughput          | Low–Medium               | High                  | Very High           |
| Process Isolation   | ✅ Full                   | ✅ Full                | ❌ Shared memory    |
| Crash Isolation     | ✅ Python crash isolated  | ✅ Isolated            | ❌ Brings down Go   |
| Memory Overhead     | Higher (extra process)   | Higher (extra service)| Lower (shared)      |
| Operational Complexity | Low                      | Medium–High           | High                |
| Best For            | ML, AI, batch jobs       | High-QPS services     | Ultra-low latency   |

## License

MIT
