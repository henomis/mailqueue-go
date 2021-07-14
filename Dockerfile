############## BUILD ############
FROM golang:alpine3.12 AS builder
#RUN go version

WORKDIR /tmp/go/
ADD go.mod /tmp/go/go.mod
RUN go mod download


######## DEV #############
FROM builder AS dev
ADD . /app
WORKDIR /app
ARG TYPE
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-w -s" -o /main ./cmd/$TYPE

######## PROD #############
FROM scratch AS prod
COPY --from=dev /main /main

ENTRYPOINT ["/main"]
