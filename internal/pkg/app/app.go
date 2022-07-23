package app

import (
	"github.com/gofiber/fiber/v2"

	"github.com/henomis/mailqueue-go/internal/pkg/audit"
	"github.com/henomis/mailqueue-go/internal/pkg/email"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemaillog"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemailqueue"
	"github.com/henomis/mailqueue-go/internal/pkg/sendmail"
)

//App struct
type App struct {
	Queue  *mongoemailqueue.MongoEmailQueue
	Log    *mongoemaillog.MongoEmailLog
	SMTP   sendmail.Client
	Server *fiber.App
}

//Options for App
type Options struct {
	Queue  *mongoemailqueue.MongoEmailQueue
	Log    *mongoemaillog.MongoEmailLog
	SMTP   sendmail.Client
	Server *fiber.App
}

//New Creates a new app instance
func New(opt Options) (*App, error) {

	app := &App{
		Server: opt.Server,
		SMTP:   opt.SMTP,
		Queue:  opt.Queue,
		Log:    opt.Log,
	}

	return app, nil
}

//RunAPI the app
func (a *App) RunAPI(address string) error {

	a.Server.Get("/api/v1/images/mail/:service/:id", a.setEmailAsRead)
	a.Server.Use("/api/v1", a.authenticationAndAuthorizationMiddleware)

	// viene chiamata dal backend per accodare un'email
	a.Server.Post("/api/v1/mail", a.enqueueEmail)
	// viene chiamata dal frontend per recuperare i dettagli di un email
	//a.Server.Get("/api/v1/mail", a.getEmailAll)
	a.Server.Get("/api/v1/mail/:id", a.getEmail)

	a.Server.Get("/api/v1/log", a.getLog)
	a.Server.Get("/api/v1/log/:id", a.getLog)

	a.Server.Get("/api/v1/template", a.template)
	a.Server.Get("/api/v1/template/:id", a.template)
	a.Server.Put("/api/v1/template/:id", a.template)
	a.Server.Post("/api/v1/template", a.template)
	a.Server.Delete("/api/v1/template/:id", a.template)

	return a.Server.Listen(address)
}

//RunPoll func
func (a *App) RunPoll() error {
	audit.Log(audit.Info, "Starting email queue poll")
	for {
		err := a.pollEmail()
		if err != nil {
			return err
		}
	}
}

func (a *App) pollEmail() error {

	dequeuedEmail, err := a.Queue.Dequeue()
	if err != nil {
		audit.Log(audit.Error, "Queue.Dequeue: %s", err.Error())
		return err
	}

	audit.Log(audit.Info, "Queue.Dequeue: %s", string(dequeuedEmail.ID))

	_, err = a.Log.Log(
		&email.Log{
			Service: dequeuedEmail.Service,
			Status:  email.StatusDequeued,
			EmailID: dequeuedEmail.ID,
		},
	)
	if err != nil {
		audit.Log(audit.Warning, "Log: %s", err.Error())
	}

	for attempt := 0; attempt < a.SMTP.Attempts(); attempt++ {
		err = a.sendEmail(dequeuedEmail)
		if err == nil {
			break
		}
	}

	if err != nil {
		audit.Log(audit.Error, "Canceled: %s", err.Error())
		errSetProcess := a.Queue.SetProcessed(dequeuedEmail.ID)
		if errSetProcess != nil {
			audit.Log(audit.Error, "Queue.SetProcessed: %s", err.Error())
		}
		a.insertLog(dequeuedEmail.ID, dequeuedEmail.Service, err.Error(), email.StatusErrorCanceled)
	}

	return nil
}

func (a *App) sendEmail(dequeuedEmail *email.Email) error {

	audit.Log(audit.Info, "Sending: %s", string(dequeuedEmail.ID))
	a.insertLog(dequeuedEmail.ID, dequeuedEmail.Service, "", email.StatusSending)

	err := a.SMTP.Send(dequeuedEmail)
	if err != nil {
		audit.Log(audit.Warning, "Send: %s, %s", string(dequeuedEmail.ID), err.Error())
		a.insertLog(dequeuedEmail.ID, dequeuedEmail.Service, err.Error(), email.StatusErrorSending)
		return err
	}

	audit.Log(audit.Info, "Send: sent %s", string(dequeuedEmail.ID))

	err = a.Queue.SetProcessed(dequeuedEmail.ID)
	if err != nil {
		audit.Log(audit.Error, "Queue.SetProcessed: %s", err.Error())
	}

	a.insertLog(dequeuedEmail.ID, dequeuedEmail.Service, "", email.StatusSent)

	return nil
}

func (a *App) insertLog(emailID, service, errorMessage string, status int) {

	_, err := a.Log.Log(
		&email.Log{
			Service: service,
			Status:  status,
			EmailID: emailID,
			Error:   errorMessage,
		},
	)
	if err != nil {
		audit.Log(audit.Warning, "Log: %s", err.Error())
	}
}

//Stop func
func (a *App) Stop() {

}
