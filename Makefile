# GoRunPy Makefile

.PHONY: all build test clean install help example

# Build and test
all: test

# Install the code generator
install:
	@echo "Installing gorunpy-gen..."
	go install ./cmd/gorunpy-gen

# Run tests (requires example to be built first)
test: example-build
	@echo "Running Go tests..."
	GORUNPY_TEST_BINARY=$(PWD)/example/dist/mathlib go test -v ./gorunpy/...

# Build and run the example
example: example-build example-run

example-build:
	@echo "Building example..."
	cd example && $(MAKE) build

example-run:
	@echo "Running example..."
	cd example && $(MAKE) run

# Clean all build artifacts
clean:
	rm -rf bin/
	cd example && $(MAKE) clean

# Format code
fmt:
	go fmt ./...

# Lint
lint:
	go vet ./...

# Help
help:
	@echo "GoRunPy - Go-native typed API for Python executables"
	@echo ""
	@echo "Usage:"
	@echo "  make install       Install gorunpy-gen to GOPATH/bin"
	@echo "  make test          Run tests (builds example first)"
	@echo "  make example       Build and run the example"
	@echo "  make example-build Build the example only"
	@echo "  make example-run   Run the example"
	@echo "  make clean         Remove build artifacts"
	@echo "  make fmt           Format Go code"
	@echo "  make lint          Lint Go code"
	@echo "  make help          Show this help"
	@echo ""
	@echo "See example/README.md for the complete example walkthrough."
