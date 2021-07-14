.PHONY: build

build:
	mkdir -p bin
	go build -o ./bin/api ./cmd/mailqueue-go-api
	go build -o ./bin/poll ./cmd/mailqueue-go-poll

test:
	go test ./... -v

.PHONY: run

#run: build
# ./bin/mystdhttp
