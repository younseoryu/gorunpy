# GoRunPy

Call Python from Go with type safety.

```go
result, _ := client.Sum(ctx, 1, 2)  // calls Python, returns 3
```

## Install

```bash
pip install gorunpy[build]
go get github.com/younseoryu/gorunpy
```

## Usage

**1. Write Python** (`py/functions.py`):
```python
from gorunpy import export

@export
def sum(a: int, b: int) -> int:
    return a + b
```

**2. Add entry point** (`py/__main__.py`):
```python
import py.functions
from gorunpy import main
if __name__ == "__main__":
    main()
```

**3. Add to your Go file**:
```go
//go:generate gorunpy build py -o .
//go:generate go run github.com/younseoryu/gorunpy/cmd/gorunpy-gen -binary py -package main -output client.go
```

**4. Build and run**:
```bash
go generate
go run .
```

## Example

```bash
cd example
python3 -m venv venv
source venv/bin/activate
pip install gorunpy[build]
go generate
go run .
```

See [example/](./example/) for complete code.
