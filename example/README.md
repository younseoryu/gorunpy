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
├── mylib/           ← Your Python code
│   ├── __init__.py
│   ├── __main__.py
│   └── functions.py
├── main.go          ← Your Go code
├── go.mod
├── dist/mylib       ← Generated binary
└── client.go        ← Generated Go client
```
