# GoRunPy Example

.PHONY: setup build generate run clean

# One-time setup
setup:
	pip install -e ../python[build]

# Build Python → executable
build:
	gorunpy build python/mathlib -o dist

# Generate Go client (run after build)
generate:
	cd .. && go run ./cmd/gorunpy-gen \
		-binary example/dist/mathlib \
		-package mathlib \
		-output example/go/mathlib/client.go

# Run
run:
	cd go && go run .

# Clean
clean:
	rm -rf dist go/mathlib/client.go
