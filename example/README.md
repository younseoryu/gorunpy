# Example

```
example/
├── py/              ← Python code
│   └── functions.py
├── py               ← Built binary (generated)
├── client.go        ← Go client (generated)
└── main.go          ← Your Go code
```

## Setup

```bash
python3 -m venv venv
source venv/bin/activate
pip install -e "../python[build]"
```

## Run

```bash
go generate
go run .
```

Output:
```
1 + 2 = 3
Hello, World!
mean = 3.0
Error: division by zero
```
