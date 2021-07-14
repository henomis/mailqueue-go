.PHONY: build

build:
	mkdir -p bin
	go build -o ./bin/api ./cmd/api
	go build -o ./bin/poll ./cmd/poll

test:
	go test ./... -v

.PHONY: run

#run: build
# ./bin/mystdhttp
