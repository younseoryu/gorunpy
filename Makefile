.PHONY: test clean

test:
	go test -v ./gorunpy/...

clean:
	rm -rf .gorunpy gorunpy_client.go
