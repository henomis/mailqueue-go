package app

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/henomis/mailqueue-go/internal/pkg/auditlogger"
	"github.com/henomis/mailqueue-go/pkg/email"
	mlog "github.com/henomis/mailqueue-go/pkg/log"
	"github.com/henomis/mailqueue-go/pkg/queue"
	"github.com/henomis/mailqueue-go/pkg/sendmail"
)

//App struct
type App struct {
	Queue       queue.Queue
	SMTP        sendmail.Client
	Server      *fiber.App
	Log         mlog.Logger
	AuditLogger auditlogger.AuditLogger
}

//Options for App
type Options struct {
	Queue       queue.Queue
	Logger      mlog.Logger
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
		Log:         opt.Logger,
		AuditLogger: opt.AuditLogger,
	}

	for {

		app.AuditLogger.Log(auditlogger.Info, "Attach queue: connecting...")
		err := opt.Queue.Attach()
		if err != nil {
			app.AuditLogger.Log(auditlogger.Error, "Attach queue: %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
		app.AuditLogger.Log(auditlogger.Info, "Attach queue: connection ok")
		break
	}

	for {

		app.AuditLogger.Log(auditlogger.Info, "Attach log: connecting...")
		err := opt.Logger.Attach()
		if err != nil {
			app.AuditLogger.Log(auditlogger.Error, "Attach queue: %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
		app.AuditLogger.Log(auditlogger.Info, "Attach log: connection ok")
		break
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

	attempts := a.SMTP.Attempts()

	for {

		e, err := a.Queue.Dequeue()
		if err != nil {
			a.AuditLogger.Log(auditlogger.Error, "Dequeue: %s", err.Error())
			return err
		}

		a.AuditLogger.Log(auditlogger.Info, "Dequeued: %s", string(e.UUID))

		entry := &mlog.Log{
			Service: e.Service,
			Status:  email.StatusDequeued,
			UUID:    e.UUID,
		}
		a.Log.Log(entry)

		for i := 0; i < attempts; i++ {

			a.AuditLogger.Log(auditlogger.Info, "Sending: %s", string(e.UUID))

			entry.Status = email.StatusSending
			entry.Error = ""
			a.Log.Log(entry)
			err = a.SMTP.Send(e)

			if err == nil {
				a.AuditLogger.Log(auditlogger.Info, "Send: sent %s", string(e.UUID))
				a.Queue.Commit(e)
				entry.Status = email.StatusSent
				a.Log.Log(entry)
				break
			}

			a.AuditLogger.Log(auditlogger.Warning, "Send: %s, %s", string(e.UUID), err.Error())
			entry.Status = email.StatusErrorSending
			entry.Error = err.Error()
			a.Log.Log(entry)

		}

		if err != nil {
			a.AuditLogger.Log(auditlogger.Error, "Canceled: %s", err.Error())
			a.Queue.Commit(e)
			entry.Status = email.StatusErrorCanceled
			entry.Error = err.Error()
			a.Log.Log(entry)
		}

	}
}

//Stop func
func (a *App) Stop() {
	a.Queue.Detach()
}
