package main

import (
	"os"
	"strconv"
	"time"

	"github.com/henomis/mailqueue-go/pkg/app"
	"github.com/henomis/mailqueue-go/pkg/limiter"
	"github.com/henomis/mailqueue-go/pkg/log"
	"github.com/henomis/mailqueue-go/pkg/queue"
	"github.com/henomis/mailqueue-go/pkg/sendmail"
	"github.com/henomis/mailqueue-go/pkg/trace"
)

type AgnosticQueue interface {
	Enqueue(interface{}) error
	Dequeue() (interface{}, error)
}

func main() {

	endpoint := os.Getenv("MONGO_ENDPOINT")
	db := os.Getenv("MONGO_DB")
	cappedSize, _ := strconv.ParseInt(os.Getenv("MONGO_DB_SIZE"), 10, 64)
	tmoI, _ := strconv.Atoi(os.Getenv("MONGO_TIMEOUT"))
	tmoD := time.Duration(tmoI) * time.Second

	allow, _ := strconv.Atoi(os.Getenv("SMTP_ALLOW"))
	interval, _ := strconv.Atoi(os.Getenv("SMTP_INTERVAL_MINUTE"))

	limit := limiter.NewDefaultLimiter(allow, time.Duration(interval)*time.Minute, &limiter.RealSleeper{})
	queue := queue.NewMongoDBQueue(queue.MongoDBOptions{Endpoint: endpoint, Database: db, CappedSize: cappedSize, Timeout: 0}, limit, nil)
	l := log.NewMongoDBLog(log.MongoDBOptions{Endpoint: endpoint, Database: db, Timeout: tmoD})
	t := trace.NewFileTracer(os.Getenv("LOG_OUTPUT"))

	clientOpt := &sendmail.Options{
		Server:   os.Getenv("SMTP_SERVER"),
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
		From:     os.Getenv("SMTP_FROM"),
		FromName: os.Getenv("SMTP_FROMNAME"),
		ReplyTo:  os.Getenv("SMTP_REPLYTO"),
		Attempts: os.Getenv("SMTP_ATTEMPTS"),
	}
	smtp := sendmail.NewMailYakClient(clientOpt)
	//smtp := sendmail.NewMockSMTPClient(clientOpt)

	opt := app.Options{
		Queue:  queue,
		Logger: l,
		Tracer: t,
		SMTP:   smtp,
	}

	poll, err := app.NewApp(opt)
	if err != nil {
		panic(err)
	}
	defer poll.Stop()

	err = poll.RunPoll()
	if err != nil {
		t.Trace(trace.Error, "RunPoll: %s", err.Error())
	}

}
