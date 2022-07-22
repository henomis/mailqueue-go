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
	"github.com/henomis/mailqueue-go/internal/pkg/auditlogger"
	fileauditlogger "github.com/henomis/mailqueue-go/internal/pkg/auditlogger/file"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemaillog"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemailqueue"

	"github.com/henomis/mailqueue-go/internal/pkg/render/mongorender"
)

func main() {

	mongoEndpoint := os.Getenv("MONGO_ENDPOINT")
	mongoDatabase := os.Getenv("MONGO_DB")
	mongoEmailDBSize, _ := strconv.ParseUint(os.Getenv("MONGO_EMAIL_DB_SIZE"), 10, 64)
	mongoLogDBSize, _ := strconv.ParseUint(os.Getenv("MONGO_LOG_DB_SIZE"), 10, 64)
	mongoTimeoutAsInt, _ := strconv.Atoi(os.Getenv("MONGO_TIMEOUT"))
	mongoTimeoutAsDuration := time.Duration(mongoTimeoutAsInt) * time.Second

	bindAddress := os.Getenv("BIND_ADDRESS")

	mongorender, err := mongorender.New(
		&mongorender.MongoRenderOptions{
			Endpoint:   mongoEndpoint,
			Database:   mongoDatabase,
			Collection: "templates",
			Timeout:    mongoTimeoutAsDuration,
		},
	)
	if err != nil {
		panic(err)
	}

	queue, err := mongoemailqueue.New(
		&mongoemailqueue.MongoEmailQueueOptions{
			Endpoint:   mongoEndpoint,
			Database:   mongoDatabase,
			Collection: "queue",
			CappedSize: mongoEmailDBSize,
			Timeout:    mongoTimeoutAsDuration,
		},
		nil,
		mongorender,
	)
	if err != nil {
		panic(err)
	}
	log, err := mongoemaillog.New(
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

	t := fileauditlogger.NewFileAuditLogger(os.Stdout)

	f := fiber.New(fiber.Config{
		StrictRouting: true,
	})
	f.Use(logger.New())
	f.Use(cors.New())
	f.Use(flimiter.New(flimiter.Config{
		Max:        200,
		Expiration: 1 * time.Minute,
	}))

	opt := app.Options{
		Log:         log,
		Queue:       queue,
		AuditLogger: t,
		Server:      f,
	}

	server, err := app.NewApp(opt)
	if err != nil {
		panic(err)
	}
	defer server.Stop()

	err = server.RunAPI(bindAddress)
	if err != nil {
		t.Log(auditlogger.Error, "RunAPI: %s", err.Error())
	}

}
