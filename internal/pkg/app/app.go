package app

import (
	"github.com/gofiber/fiber/v2"

	"github.com/henomis/mailqueue-go/internal/pkg/auditlogger"
	"github.com/henomis/mailqueue-go/internal/pkg/email"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemaillog"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemailqueue"
	"github.com/henomis/mailqueue-go/internal/pkg/sendmail"
)

//App struct
type App struct {
	Queue       *mongoemailqueue.MongoEmailQueue
	Log         *mongoemaillog.MongoEmailLog
	SMTP        sendmail.Client
	Server      *fiber.App
	AuditLogger auditlogger.AuditLogger
}

//Options for App
type Options struct {
	Queue       *mongoemailqueue.MongoEmailQueue
	Log         *mongoemaillog.MongoEmailLog
	AuditLogger auditlogger.AuditLogger
	SMTP        sendmail.Client
	Server      *fiber.App
}

//NewApp Creates a new app instance
func NewApp(opt Options) (*App, error) {

	app := &App{
		Server:      opt.Server,
		SMTP:        opt.SMTP,
		Queue:       opt.Queue,
		Log:         opt.Log,
		AuditLogger: opt.AuditLogger,
	}

	return app, nil
}

//RunAPI the app
func (a *App) RunAPI(address string) error {
	a.Server.Get("/img/mail/:uuid", a.readEmail)

	a.Server.Use("/api/v1", a.authenticationAndAuthorizationMiddleware)

	// viene chiamata dal backend per accodare un'email
	a.Server.Post("/api/v1/mail", a.enqueueEmail)
	// viene chiamata dal frontend per recuperare i dettagli di un email
	//a.Server.Get("/api/v1/mail", a.getEmailAll)
	a.Server.Get("/api/v1/mail/:uuid", a.getEmail)

	a.Server.Get("/api/v1/log", a.getLog)
	a.Server.Get("/api/v1/log/:uuid", a.getLog)

	a.Server.Get("/api/v1/template", a.template)
	a.Server.Get("/api/v1/template/:id", a.template)
	a.Server.Put("/api/v1/template/:id", a.template)
	a.Server.Post("/api/v1/template", a.template)
	a.Server.Delete("/api/v1/emplate/:id", a.template)

	return a.Server.Listen(address)
}

//RunPoll func
func (a *App) RunPoll() error {

	for {

		dequeuedEmail, err := a.Queue.Dequeue()
		if err != nil {
			a.AuditLogger.Log(auditlogger.Error, "Dequeue: %s", err.Error())
			return err
		}

		a.AuditLogger.Log(auditlogger.Info, "Dequeued: %s", string(dequeuedEmail.ID))

		entry := &email.Log{
			Service: dequeuedEmail.Service,
			Status:  email.StatusDequeued,
			EmailID: dequeuedEmail.ID,
		}
		a.Log.Log(entry)

		for attempt := 0; attempt < a.SMTP.Attempts(); attempt++ {

			a.AuditLogger.Log(auditlogger.Info, "Sending: %s", string(dequeuedEmail.ID))

			entry.Status = email.StatusSending
			entry.Error = ""
			a.Log.Log(entry)
			err = a.SMTP.Send(dequeuedEmail)

			if err == nil {
				a.AuditLogger.Log(auditlogger.Info, "Send: sent %s", string(dequeuedEmail.ID))
				a.Queue.SetProcessed(dequeuedEmail.ID)
				entry.Status = email.StatusSent
				a.Log.Log(entry)
				break
			}

			a.AuditLogger.Log(auditlogger.Warning, "Send: %s, %s", string(dequeuedEmail.ID), err.Error())
			entry.Status = email.StatusErrorSending
			entry.Error = err.Error()
			a.Log.Log(entry)

		}

		if err != nil {
			a.AuditLogger.Log(auditlogger.Error, "Canceled: %s", err.Error())
			a.Queue.SetProcessed(dequeuedEmail.ID)
			entry.Status = email.StatusErrorCanceled
			entry.Error = err.Error()
			a.Log.Log(entry)
		}

	}
}

//Stop func
func (a *App) Stop() {

}
