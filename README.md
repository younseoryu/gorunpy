# GoRunPy

Call Python from Go with type safety.

```go
result, _ := client.Sum(ctx, 1, 2)  // calls Python, returns 3
```

## Install

```bash
go get github.com/younseoryu/gorunpy
pip install gorunpy[build]
```

## Usage

**1. Write Python:**
```python
from gorunpy import export

@export
def sum(a: int, b: int) -> int:
    return a + b
```

**2. Build:**
```bash
gorunpy build ./mymodule -o ./dist
```

**3. Generate Go client:**
```bash
gorunpy-gen -binary ./dist/mymodule -package mymodule -output mymodule/client.go
```

**4. Use:**
```go
client := mymodule.NewClient("./dist/mymodule")
result, _ := client.Sum(ctx, 1, 2)
```

## Example

See [example/](./example/) for a complete working example.
