######## BUILD #############
FROM golang:alpine3.15 AS builder
WORKDIR /app
ADD . /app
RUN go mod download
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-w -s" -o /api ./cmd/api

######## PROD #############
FROM scratch AS prod
COPY --from=builder /api /api

ENTRYPOINT ["/api"]
