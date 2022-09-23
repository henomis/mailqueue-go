######## BUILD #############
FROM golang:alpine3.15 AS builder
WORKDIR /app
ADD . /app
RUN go mod download
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-w -s" -o /watcher ./cmd/watcher

######## PROD #############
FROM scratch AS prod
COPY --from=builder /watcher /watcher

ENTRYPOINT ["/watcher"]
