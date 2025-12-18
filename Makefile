.PHONY: test example clean

test:
	cd example && make build && make generate
	GORUNPY_TEST_BINARY=$(PWD)/example/dist/mathlib go test -v ./gorunpy/...

example:
	cd example && make build && make generate && make run

clean:
	cd example && make clean
