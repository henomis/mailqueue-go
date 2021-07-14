package app

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/henomis/mailqueue-go/pkg/email"
	"github.com/henomis/mailqueue-go/pkg/log"
	mlog "github.com/henomis/mailqueue-go/pkg/log"
	"github.com/henomis/mailqueue-go/pkg/queue"
	"github.com/henomis/mailqueue-go/pkg/sendmail"
	"github.com/henomis/mailqueue-go/pkg/trace"
)

//App struct
type App struct {
	Queue  queue.Queue
	SMTP   sendmail.Client
	Server *fiber.App
	Log    mlog.Logger
	Tracer trace.Tracer
}

//Options for App
type Options struct {
	Queue  queue.Queue
	Logger mlog.Logger
	Tracer trace.Tracer
	SMTP   sendmail.Client
	Server *fiber.App
}

//NewApp Creates a new app instance
func NewApp(opt Options) (*App, error) {

	app := &App{
		Server: opt.Server,
		SMTP:   opt.SMTP,
		Queue:  opt.Queue,
		Log:    opt.Logger,
		Tracer: opt.Tracer,
	}

	for {

		app.Tracer.Trace(trace.Info, "Attach queue: connecting...")
		err := opt.Queue.Attach()
		if err != nil {
			app.Tracer.Trace(trace.Error, "Attach queue: %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
		app.Tracer.Trace(trace.Info, "Attach queue: connection ok")
		break
	}

	for {

		app.Tracer.Trace(trace.Info, "Attach log: connecting...")
		err := opt.Logger.Attach()
		if err != nil {
			app.Tracer.Trace(trace.Error, "Attach queue: %s", err)
			time.Sleep(1 * time.Second)
			continue
		}
		app.Tracer.Trace(trace.Info, "Attach log: connection ok")
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
			a.Tracer.Trace(trace.Error, "Dequeue: %s", err.Error())
			return err
		}

		a.Tracer.Trace(trace.Info, "Dequeued: %s", string(e.UUID))

		entry := &log.Log{
			Service: e.Service,
			Status:  email.StatusDequeued,
			UUID:    e.UUID,
		}
		a.Log.Log(entry)

		for i := 0; i < attempts; i++ {

			a.Tracer.Trace(trace.Info, "Sending: %s", string(e.UUID))

			entry.Status = email.StatusSending
			entry.Error = ""
			a.Log.Log(entry)
			err = a.SMTP.Send(e)

			if err == nil {
				a.Tracer.Trace(trace.Info, "Send: sent %s", string(e.UUID))
				a.Queue.Commit(e)
				entry.Status = email.StatusSent
				a.Log.Log(entry)
				break
			}

			a.Tracer.Trace(trace.Warning, "Send: %s, %s", string(e.UUID), err.Error())
			entry.Status = email.StatusErrorSending
			entry.Error = err.Error()
			a.Log.Log(entry)

		}

		if err != nil {
			a.Tracer.Trace(trace.Error, "Canceled: %s", err.Error())
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
