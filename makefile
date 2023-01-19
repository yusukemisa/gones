build:
	go build -o bin/gones

clean:
	rm -f ./bin/gones
.PHONY: clean

test:
	go test -v -race ./...
.PHONY: test
