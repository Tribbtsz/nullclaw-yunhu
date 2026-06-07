.PHONY: build test lint clean

build:
	go build -o yunhu-channel ./cmd/yunhu-channel

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -f yunhu-channel
