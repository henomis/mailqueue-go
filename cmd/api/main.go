package main

import (
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	flimiter "github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/henomis/mailqueue-go/internal/pkg/app"
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

	mongotemplate, err := mongotemplate.New(
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
		mongotemplate,
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

	httpServer := fiber.New(fiber.Config{
		StrictRouting: true,
	})
	httpServer.Use(logger.New())
	httpServer.Use(cors.New())
	httpServer.Use(flimiter.New(flimiter.Config{
		Max:        200,
		Expiration: 1 * time.Minute,
	}))

	appOptions := app.AppOptions{
		EmailLog:      mongoEmailLog,
		EmailQueue:    mongoEmailQueue,
		EmailTemplate: mongotemplate,
		HTTPServer:    httpServer,
	}

	server, err := app.New(appOptions)
	if err != nil {
		panic(err)
	}
	defer server.Stop()

	err = server.RunAPI(bindAddress)
	if err != nil {
		audit.Log(audit.Error, "RunAPI: %s", err.Error())
	}

}
