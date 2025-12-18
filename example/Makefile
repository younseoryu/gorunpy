.PHONY: setup build generate run clean

setup:
	pip install -e ../python[build]

build:
	gorunpy build python/mathlib -o dist

generate:
	cd .. && go run ./cmd/gorunpy-gen -binary example/dist/mathlib -package mathlib -output example/go/mathlib/client.go

run:
	cd go && go run .

clean:
	rm -rf dist go/mathlib/client.go
