# 📤 Mailqueue-go

General purpose email queue with REST API interface.

## Features

- 📎 Full attachment support (via base64 encoding or filesystem)
- 👨‍👩‍👧‍👦 Full suppor to Cc, Bcc, reply-to fields
- 🌐 HTML template engine
- 🔒 Automatic TLS support
- ✉️ SMTP client send limiter

## Build

Use `Makefile` to build `bin/mailqueue-go-api` and `bin/mailqueue-go-watcher`.

```bash
make
```


## Docker

You can start a complete mailqueue-go stack with docker-compose.

```bash
docker-compose up
```

## Env

Put this in your `.env` file and modify it to your needs.

```bash
MONGO_ENDPOINT=mongodb://admin:pass@mongodb:27017
MONGO_DB=test
MONGO_LOG_DB_SIZE=1000000
MONGO_EMAIL_DB_SIZE=1000000
MONGO_TIMEOUT=10
BIND_ADDRESS=:8080
SMTP_ALLOW=10
SMTP_INTERVAL_MINUTE=1
SMTP_SERVER=localhost
SMTP_USERNAME=username@localhost
SMTP_PASSWORD=password
SMTP_FROM=username@localhost
SMTP_FROMNAME=fromname
SMTP_REPLYTO=noreply@localhost
SMTP_ATTEMPTS=3
LOG_OUTPUT=stdout
````

## API

API endpoint prefix is `/api/v1`.

| Method | Route | Description |
|--------|-------|-------------|
|`GET`| `/logs` | Get all logs|
|`GET`| `/logs/{email_id}`|  Get logs for email with id `{email_id}`|
|`GET`| `/emails` | Get all emails|
|`GET`| `/emails/{id}` | Get email with id `{id}`|
|`POST`| `/emails` | Enqueue new email|
|`GET`| `/templates` | Get all templates|
|`GET`| `/templates/{id}` | Get template with id `{id}`|
|`PUT`| `/templates/{id}` | Update template with id `{id}`|
|`POST`| `/templates` | Create new template|
|`DELETE`| `/templates/{id}` | Delete template with id `{id}`|
|`GET`|`/images/mail/{service}/{id}`| Tracking open email|

