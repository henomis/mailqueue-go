.PHONY: build

build:
	mkdir -p bin
	go build -o ./bin/mailqueue-go-api ./cmd/api
	go build -o ./bin/mailqueue-go-watcher ./cmd/watcher

clean:
	rm -fr ./bin

test:
	go test ./... -v


