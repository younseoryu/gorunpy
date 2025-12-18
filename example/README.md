# GoRunPy Example

## Setup (once)

```bash
pip install -e ../python[build]
```

## Build & Run

```bash
make build      # Python → executable
make generate   # Generate Go client
make run        # Run Go program
```

## Output

```
1 + 2 = 3
3.5 * 2.0 = 7.0
Hello, World!
mean([1,2,3,4,5]) = 3.0
Error: division by zero
```

## Files

```
python/mathlib/functions.py   ← You write this
dist/mathlib                  ← Generated executable  
go/mathlib/client.go          ← Generated Go client
go/main.go                    ← You write this
```
