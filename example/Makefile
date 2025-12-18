.PHONY: setup generate run clean

setup:
	pip install -e "../python[build]"

generate:
	go generate

run:
	go run .

clean:
	rm -rf .gorunpy gorunpy_client.go
