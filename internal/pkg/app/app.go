package app

import (
	"github.com/gofiber/fiber/v2"

	"github.com/henomis/mailqueue-go/internal/pkg/audit"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemaillog"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemailqueue"
	"github.com/henomis/mailqueue-go/internal/pkg/mongotemplate"
	"github.com/henomis/mailqueue-go/internal/pkg/sendmail"
	"github.com/henomis/mailqueue-go/internal/pkg/storagemodel"
)

//App struct
type App struct {
	emailQueue    *mongoemailqueue.MongoEmailQueue
	emailLog      *mongoemaillog.MongoEmailLog
	emailTemplate *mongotemplate.MongoTemplate
	smtpClient    sendmail.Client
	httpServer    *fiber.App
}

//AppOptions for App
type AppOptions struct {
	EmailQueue    *mongoemailqueue.MongoEmailQueue
	EmailLog      *mongoemaillog.MongoEmailLog
	EmailTemplate *mongotemplate.MongoTemplate
	SMTPClient    sendmail.Client
	HTTPServer    *fiber.App
}

//New Creates a new app instance
func New(appOptions AppOptions) (*App, error) {

	app := &App{
		httpServer:    appOptions.HTTPServer,
		smtpClient:    appOptions.SMTPClient,
		emailQueue:    appOptions.EmailQueue,
		emailLog:      appOptions.EmailLog,
		emailTemplate: appOptions.EmailTemplate,
	}

	return app, nil
}

//RunAPI the app
func (a *App) RunAPI(address string) error {

	a.httpServer.Get("/api/v1/images/mail/:service/:id", a.setEmailAsRead)
	a.httpServer.Use("/api/v1", a.authenticationAndAuthorizationMiddleware)

	// LOGS
	a.httpServer.Get("/api/v1/logs", a.getLogs)
	a.httpServer.Get("/api/v1/logs/:email_id", a.getLog)

	// EMAILS
	a.httpServer.Get("/api/v1/emails", a.getEmails)
	a.httpServer.Get("/api/v1/emails/:id", a.getEmail)
	a.httpServer.Post("/api/v1/emails", a.enqueueEmail)

	// TEMPLATES
	a.httpServer.Get("/api/v1/templates", a.getTemplates)
	a.httpServer.Get("/api/v1/templates/:id", a.getTemplate)
	a.httpServer.Put("/api/v1/templates/:id", a.updateTemplate)
	a.httpServer.Post("/api/v1/templates", a.addTemplate)
	a.httpServer.Delete("/api/v1/templates/:id", a.deleteTemplate)

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
	a.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, "", storagemodel.StatusDequeued)

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
		a.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, err.Error(), storagemodel.StatusErrorCanceled)
	}

	return nil
}

func (a *App) sendEmail(dequeuedEmail *storagemodel.Email) error {

	audit.Log(audit.Info, "Sending: %s", string(dequeuedEmail.ID))
	a.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, "", storagemodel.StatusSending)

	err := a.smtpClient.Send(dequeuedEmail)
	if err != nil {
		audit.Log(audit.Warning, "Send: %s, %s", string(dequeuedEmail.ID), err.Error())
		a.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, err.Error(), storagemodel.StatusErrorSending)
		return err
	}

	audit.Log(audit.Info, "Send: sent %s", string(dequeuedEmail.ID))

	err = a.emailQueue.SetProcessed(dequeuedEmail.ID)
	if err != nil {
		audit.Log(audit.Error, "Queue.SetProcessed: %s", err.Error())
	}

	a.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, "", storagemodel.StatusSent)

	return nil
}

func (a *App) addEmailLog(emailID, service, errorMessage string, status int) {

	_, err := a.emailLog.Log(
		&storagemodel.Log{
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
