package main

import (
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	flimiter "github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/henomis/mailqueue-go/pkg/app"
	"github.com/henomis/mailqueue-go/pkg/log"
	"github.com/henomis/mailqueue-go/pkg/queue"
	"github.com/henomis/mailqueue-go/pkg/render"
	"github.com/henomis/mailqueue-go/pkg/trace"
)

func main() {

	endpoint := os.Getenv("MONGO_ENDPOINT")
	db := os.Getenv("MONGO_DB")
	cappedSize, _ := strconv.ParseInt(os.Getenv("MONGO_DB_SIZE"), 10, 64)
	timeoutI, _ := strconv.Atoi(os.Getenv("MONGO_TIMEOUT"))
	timeoutD := time.Duration(timeoutI) * time.Second

	bindAddress := os.Getenv("BIND_ADDRESS")

	tmpl := render.NewMongoRender(timeoutD, endpoint, db)
	q := queue.NewMongoDBQueue(queue.MongoDBOptions{Endpoint: endpoint, Database: db, CappedSize: cappedSize, Timeout: timeoutD}, nil, tmpl)
	l := log.NewMongoDBLog(log.MongoDBOptions{Endpoint: endpoint, Database: db, Timeout: timeoutD})
	t := trace.NewFileTracer(os.Getenv("LOG_OUTPUT"))

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
		Logger: l,
		Queue:  q,
		Tracer: t,
		Server: f,
	}

	server, err := app.NewApp(opt)
	if err != nil {
		panic(err)
	}
	defer server.Stop()

	err = server.RunAPI(bindAddress)
	if err != nil {
		t.Trace(trace.Error, "RunAPI: %s", err.Error())
	}

}
