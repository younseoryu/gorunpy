# GoRunPy Example

Standalone example. Copy this folder anywhere and run.

## Setup (once)

```bash
python3 -m venv venv
source venv/bin/activate
pip install "gorunpy[build]"
```

## Run

```bash
source venv/bin/activate
go generate
go run .
```

## Output

```
1 + 2 = 3
Hello, World!
mean = 3.0
Error: division by zero
```

## Files

```
example/
├── mylib/               ← Your Python code
│   ├── __init__.py
│   └── functions.py     ← @gorunpy.export functions
├── main.go              ← //go:generate gorunpy gen
├── gorunpy_client.go    ← Generated Go client (with embedded binary)
└── .gorunpy/            ← Hidden build artifacts
    └── mylib            ← Compiled binary
```

## What `gorunpy gen` does

1. Auto-detects `mylib/` (has `@gorunpy.export`)
2. Builds binary → `.gorunpy/mylib`
3. Generates `gorunpy_client.go` with:
   - `//go:embed` directive
   - `NewClient()` function
   - Type-safe method wrappers
