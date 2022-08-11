package main

import (
	"os"
	"strconv"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/app/httpserver"
	"github.com/henomis/mailqueue-go/internal/pkg/audit"
	"github.com/henomis/mailqueue-go/internal/pkg/mongotemplate"

	"github.com/henomis/mailqueue-go/internal/pkg/mongoemaillog"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemailqueue"
)

func main() {

	mongoEndpoint := os.Getenv("MONGO_ENDPOINT")
	mongoDatabase := os.Getenv("MONGO_DB")
	mongoEmailDBSize, _ := strconv.ParseUint(os.Getenv("MONGO_EMAIL_DB_SIZE"), 10, 64)
	mongoLogDBSize, _ := strconv.ParseUint(os.Getenv("MONGO_LOG_DB_SIZE"), 10, 64)
	mongoTimeoutAsInt, _ := strconv.Atoi(os.Getenv("MONGO_TIMEOUT"))
	mongoTimeoutAsDuration := time.Duration(mongoTimeoutAsInt) * time.Second

	bindAddress := os.Getenv("BIND_ADDRESS")

	mongoTemplate, err := mongotemplate.New(
		&mongotemplate.MongoTemplateOptions{
			Endpoint:   mongoEndpoint,
			Database:   mongoDatabase,
			Collection: "templates",
			Timeout:    mongoTimeoutAsDuration,
		},
	)
	if err != nil {
		panic(err)
	}

	mongoEmailQueue, err := mongoemailqueue.New(
		&mongoemailqueue.MongoEmailQueueOptions{
			Endpoint:   mongoEndpoint,
			Database:   mongoDatabase,
			Collection: "queue",
			CappedSize: mongoEmailDBSize,
			Timeout:    mongoTimeoutAsDuration,
		},
		nil,
		mongoTemplate,
	)
	if err != nil {
		panic(err)
	}
	mongoEmailLog, err := mongoemaillog.New(
		&mongoemaillog.MongoEmailLogOptions{
			Endpoint:   mongoEndpoint,
			Database:   mongoDatabase,
			Collection: "log",
			CappedSize: mongoLogDBSize,
			Timeout:    mongoTimeoutAsDuration,
		},
	)
	if err != nil {
		panic(err)
	}

	httpServer := httpserver.New(
		mongoEmailQueue,
		mongoEmailLog,
		mongoTemplate,
	)

	err = httpServer.Run(bindAddress)
	if err != nil {
		audit.Log(audit.Error, "httpServer.Run: %s", err.Error())
	}

}
