package watcher

import (
	"github.com/henomis/mailqueue-go/internal/pkg/app"
	"github.com/henomis/mailqueue-go/internal/pkg/audit"
	"github.com/henomis/mailqueue-go/internal/pkg/storagemodel"
)

type SendmailClient interface {
	Send(e *storagemodel.Email) error
	Attempts() int
}

type Watcher struct {
	smtpClient SendmailClient
	emailQueue app.EmailQueue
	emailLog   app.EmailLog
}

func New(
	smtpClient SendmailClient,
	emailQueue app.EmailQueue,
	emailLog app.EmailLog,
) *Watcher {
	return &Watcher{
		smtpClient: smtpClient,
		emailQueue: emailQueue,
		emailLog:   emailLog,
	}
}

func (w *Watcher) Run() error {

	audit.Log(audit.Info, "Starting email queue watcher")
	for {
		err := w.watchEmail()
		if err != nil {
			return err
		}
	}

}

func (w *Watcher) watchEmail() error {

	dequeuedEmail, err := w.emailQueue.Dequeue()
	if err != nil {
		audit.Log(audit.Error, "Queue.Dequeue: %s", err.Error())
		return err
	}

	audit.Log(audit.Info, "Queue.Dequeue: %s", string(dequeuedEmail.ID))
	w.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, "", storagemodel.StatusDequeued)

	for attempt := 0; attempt < w.smtpClient.Attempts(); attempt++ {
		err = w.sendEmail(dequeuedEmail)
		if err == nil {
			break
		}
	}

	if err != nil {
		audit.Log(audit.Error, "Canceled: %s", err.Error())
		errSetProcess := w.emailQueue.SetProcessed(dequeuedEmail.ID)
		if errSetProcess != nil {
			audit.Log(audit.Error, "Queue.SetProcessed: %s", err.Error())
		}
		w.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, err.Error(), storagemodel.StatusErrorCanceled)
	}

	return nil
}

func (w *Watcher) sendEmail(dequeuedEmail *storagemodel.Email) error {

	audit.Log(audit.Info, "Sending: %s", string(dequeuedEmail.ID))
	w.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, "", storagemodel.StatusSending)

	err := w.smtpClient.Send(dequeuedEmail)
	if err != nil {
		audit.Log(audit.Warning, "Send: %s, %s", string(dequeuedEmail.ID), err.Error())
		w.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, err.Error(), storagemodel.StatusErrorSending)
		return err
	}

	audit.Log(audit.Info, "Send: sent %s", string(dequeuedEmail.ID))

	err = w.emailQueue.SetProcessed(dequeuedEmail.ID)
	if err != nil {
		audit.Log(audit.Error, "Queue.SetProcessed: %s", err.Error())
	}

	w.addEmailLog(dequeuedEmail.ID, dequeuedEmail.Service, "", storagemodel.StatusSent)

	return nil
}

func (w *Watcher) addEmailLog(emailID, service, errorMessage string, status int) {

	_, err := w.emailLog.Create(
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
