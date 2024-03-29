package main

import (
	"os"
	"strconv"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/app/watcher"
	"github.com/henomis/mailqueue-go/internal/pkg/audit"
	"github.com/henomis/mailqueue-go/internal/pkg/limiter"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemaillog"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemailqueue"
	"github.com/henomis/mailqueue-go/internal/pkg/sendmail/mailyakclient"
)

func main() {

	mongoEndpoint := os.Getenv("MONGO_ENDPOINT")
	mongoDatabase := os.Getenv("MONGO_DB")
	mongoEmailDBSize, _ := strconv.ParseUint(os.Getenv("MONGO_EMAIL_DB_SIZE"), 10, 64)
	mongoLogDBSize, _ := strconv.ParseUint(os.Getenv("MONGO_LOG_DB_SIZE"), 10, 64)
	mongoTimeoutAsInt, _ := strconv.Atoi(os.Getenv("MONGO_TIMEOUT"))
	mongoTimeoutAsDuration := time.Duration(mongoTimeoutAsInt) * time.Second

	limiterAllowed, _ := strconv.ParseUint(os.Getenv("SMTP_ALLOW"), 10, 64)
	limiterInterval, _ := strconv.Atoi(os.Getenv("SMTP_INTERVAL_MINUTE"))

	fixedWindowLiminter := limiter.NewFixedWindowLimiter(limiterAllowed, time.Duration(limiterInterval)*time.Minute)

	mongoEmailQueue, err := mongoemailqueue.New(
		&mongoemailqueue.MongoEmailQueueOptions{
			Endpoint:   mongoEndpoint,
			Database:   mongoDatabase,
			Collection: "queue",
			CappedSize: mongoEmailDBSize,
			Timeout:    mongoTimeoutAsDuration,
		},
		fixedWindowLiminter,
		nil,
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

	smtpClient := mailyakclient.New(
		&mailyakclient.MailYakClientOptions{
			Server:   os.Getenv("SMTP_SERVER"),
			Username: os.Getenv("SMTP_USERNAME"),
			Password: os.Getenv("SMTP_PASSWORD"),
			From:     os.Getenv("SMTP_FROM"),
			FromName: os.Getenv("SMTP_FROMNAME"),
			ReplyTo:  os.Getenv("SMTP_REPLYTO"),
			Attempts: os.Getenv("SMTP_ATTEMPTS"),
		},
	)

	watcher := watcher.New(
		smtpClient,
		mongoEmailQueue,
		mongoEmailLog,
	)

	err = watcher.Run()
	if err != nil {
		audit.Log(audit.Error, "watcher.Run: %s", err.Error())
	}

}
