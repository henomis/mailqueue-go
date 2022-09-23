.PHONY: build

build:
	mkdir -p bin
	go build -o ./bin/mailqueue-go-api ./cmd/api
	go build -o ./bin/mailqueue-go-poll ./cmd/poll

test:
	go test ./... -v


