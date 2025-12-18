.PHONY: test example clean

test:
	cd example && make generate
	go test -v ./gorunpy/...

example:
	cd example && make generate && make run

clean:
	cd example && make clean
