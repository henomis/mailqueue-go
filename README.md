<!-- TOC -->

- [Features](#features)
- [mailqueue-go-api](#mailqueue-go-api)
    - [email](#email)
    - [log](#log)
    - [template](#template)
    - [read](#read)
- [mailqueue-go-poll](#mailqueue-go-poll)
- [Build](#build)
- [Docker](#docker)
- [Env](#env)
    - [api](#api)
    - [poll](#poll)

<!-- /TOC -->

Simple SMTP HTTP/API client that uses MongoDB as queue.

## Features

- Full attachment support (via base64 encoding or filesystem)
- Full suppor to Cc, Bcc, reply-to fields
- HTML template engine
- Automatic TLS support
- SMTP client send limiter

## mailqueue-go-api

This is the API HTTP REST backend. You can use the following endpoints/verbs:

### email

This enqueue or get email resources.

```bash
POST /api/v1/mail
GET /api/v1/mail
GET /api/v1/mail/:uuid
```

### log

This will retreive events related to one email

```bash
GET /api/v1/log
GET /api/v1/log/:uuid
```

### template

These are CRUD operation over templates

```bash
GET /api/v1/template
GET /api/v1/template/:id
PUT /api/v1/template/:id
POST /api/v1/template
DELETE /api/v1/emplate/:id
```

### read

This will support white pixel and marks as read email.

```bash
GET /img/mail/:uuid
```

## mailqueue-go-poll

This is the SMTP client, it will refer to one email account service.

## Build

This command will build `bin/mailqueue-go-api` and `bin/mailqueue-go-poll`.

```bash
make
```

## Docker

If you are using Docker you can start the whole system with:

```bash
docker-compose up
```

## Env

You have to setup variables in yout bash environnment or in your `docker-compose.yml`

### api

```bash
export MONGO_ENDPOINT=mongodb://admin:pass@localhost:27017/admin?authSource=admin
export MONGO_DB=test
export MONGO_DB_SIZE=1000000
export MONGO_TIMEOUT=10
export BIND_ADDRESS=":8080"
export LOG_OUTPUT="stdout"
```

### poll

```bash
export MONGO_ENDPOINT=mongodb://admin:pass@localhost:27017/admin?authSource=admin
export MONGO_DB=test
export MONGO_DB_SIZE=1000000
export MONGO_TIMEOUT=10
export SMTP_ALLOW=10
export SMTP_INTERVAL_MINUTE=1
export SMTP_SERVER=localhost
export SMTP_USERNAME=username@localhost
export SMTP_PASSWORD=password
export SMTP_FROM=username@localhost
export SMTP_FROMNAME=fromname
export SMTP_REPLYTO=noreply@localhost
export SMTP_ATTEMPTS=3
export LOG_OUTPUT=stdout
```
