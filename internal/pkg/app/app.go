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
	emailQueue *mongoemailqueue.MongoEmailQueue
	emailLog   *mongoemaillog.MongoEmailLog
	smtpClient sendmail.Client
	httpServer *fiber.App
}

//AppOptions for App
type AppOptions struct {
	EmailQueue *mongoemailqueue.MongoEmailQueue
	EmailLog   *mongoemaillog.MongoEmailLog
	SMTPClient sendmail.Client
	HTTPServer *fiber.App
}

//New Creates a new app instance
func New(appOptions AppOptions) (*App, error) {

	app := &App{
		httpServer: appOptions.HTTPServer,
		smtpClient: appOptions.SMTPClient,
		emailQueue: appOptions.EmailQueue,
		emailLog:   appOptions.EmailLog,
	}

	return app, nil
}

//RunAPI the app
func (a *App) RunAPI(address string) error {

	a.httpServer.Get("/api/v1/images/mail/:service/:id", a.setEmailAsRead)
	a.httpServer.Use("/api/v1", a.authenticationAndAuthorizationMiddleware)

	// viene chiamata dal backend per accodare un'email
	a.httpServer.Post("/api/v1/mail", a.enqueueEmail)
	// viene chiamata dal frontend per recuperare i dettagli di un email
	//a.Server.Get("/api/v1/mail", a.getEmailAll)
	a.httpServer.Get("/api/v1/mail/:id", a.getEmail)

	a.httpServer.Get("/api/v1/log", a.getLog)
	a.httpServer.Get("/api/v1/log/:email_id", a.getLog)

	a.httpServer.Get("/api/v1/template", a.template)
	a.httpServer.Get("/api/v1/template/:id", a.template)
	a.httpServer.Put("/api/v1/template/:id", a.template)
	a.httpServer.Post("/api/v1/template", a.template)
	a.httpServer.Delete("/api/v1/template/:id", a.template)

	return a.httpServer.Listen(address)
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

	dequeuedEmail, err := a.emailQueue.Dequeue()
	if err != nil {
		audit.Log(audit.Error, "Queue.Dequeue: %s", err.Error())
		return err
	}

	audit.Log(audit.Info, "Queue.Dequeue: %s", string(dequeuedEmail.ID))
	a.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, "", email.StatusDequeued)

	for attempt := 0; attempt < a.smtpClient.Attempts(); attempt++ {
		err = a.sendEmail(dequeuedEmail)
		if err == nil {
			break
		}
	}

	if err != nil {
		audit.Log(audit.Error, "Canceled: %s", err.Error())
		errSetProcess := a.emailQueue.SetProcessed(dequeuedEmail.ID)
		if errSetProcess != nil {
			audit.Log(audit.Error, "Queue.SetProcessed: %s", err.Error())
		}
		a.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, err.Error(), email.StatusErrorCanceled)
	}

	return nil
}

func (a *App) sendEmail(dequeuedEmail *email.Email) error {

	audit.Log(audit.Info, "Sending: %s", string(dequeuedEmail.ID))
	a.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, "", email.StatusSending)

	err := a.smtpClient.Send(dequeuedEmail)
	if err != nil {
		audit.Log(audit.Warning, "Send: %s, %s", string(dequeuedEmail.ID), err.Error())
		a.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, err.Error(), email.StatusErrorSending)
		return err
	}

	audit.Log(audit.Info, "Send: sent %s", string(dequeuedEmail.ID))

	err = a.emailQueue.SetProcessed(dequeuedEmail.ID)
	if err != nil {
		audit.Log(audit.Error, "Queue.SetProcessed: %s", err.Error())
	}

	a.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, "", email.StatusSent)

	return nil
}

func (a *App) addEmailLog(emailID, service, errorMessage string, status int) {

	_, err := a.emailLog.Log(
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
