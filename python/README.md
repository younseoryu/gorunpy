# GoRunPy - Python SDK

Python SDK for creating Go-callable typed executables.

## Installation

```bash
pip install gorunpy
```

For building executables with PyInstaller:

```bash
pip install gorunpy[build]
```

## Quick Start

### 1. Define Exported Functions

```python
# mymodule/functions.py
from gorunpy import export, ValidationError

@export
def add(a: int, b: int) -> int:
    """Add two integers."""
    return a + b

@export
def divide(a: float, b: float) -> float:
    """Divide two numbers."""
    if b == 0:
        raise ValidationError("division by zero", field="b")
    return a / b
```

### 2. Create Entry Point

```python
# mymodule/__main__.py
import mymodule.functions  # Register functions
from gorunpy import main

if __name__ == "__main__":
    main()
```

### 3. Build Executable

```bash
pyinstaller --onefile --name mymodule mymodule/__main__.py
```

### 4. Call from Go

```go
import "github.com/user/gorunpy/gorunpy"

client := gorunpy.NewClient("./dist/mymodule")
result, err := client.CallRaw(ctx, "add", map[string]any{"a": 1, "b": 2})
```

## Type Annotations

All exported functions must have complete type annotations:

```python
@export
def process(
    name: str,
    count: int,
    values: List[float],
    options: Optional[Dict[str, bool]] = None
) -> Dict[str, Any]:
    ...
```

Supported types:
- `int`, `float`, `str`, `bool`, `None`
- `List[T]`, `Dict[str, T]`
- `Optional[T]`, `Union[T1, T2, ...]`
- `Any`

## Error Handling

```python
from gorunpy import ValidationError, GoRunPyError

# Validation error (exit code 1)
raise ValidationError("invalid value", field="my_field")

# Custom error (exit code 1)
raise GoRunPyError("something went wrong", kind="CustomError")

# Unhandled exceptions become RuntimeError (exit code 2)
```

## Protocol

Request (stdin):
```json
{"function": "add", "args": {"a": 1, "b": 2}}
```

Success response (stdout):
```json
{"ok": true, "result": {"value": 3}}
```

Error response (stderr):
```json
{"ok": false, "error": {"kind": "ValidationError", "message": "...", "field": "..."}}
```

## License

MIT

