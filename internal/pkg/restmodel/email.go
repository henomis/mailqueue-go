package restmodel

import (
	"github.com/henomis/mailqueue-go/internal/pkg/storagemodel"
)

type Emails []Email

type Email struct {
	ID          string      `json:"id"`
	Service     string      `json:"service"`
	To          string      `json:"to"`
	Cc          string      `json:"cc"`
	Bcc         string      `json:"bcc"`
	Subject     string      `json:"subject"`
	Data        string      `json:"data"`
	HTML        string      `json:"html"`
	Template    string      `json:"template"`
	Attachments Attachments `json:"attachments"`
	Log         Logs        `json:"log,omitempty"`
}

type Attachments []Attachment
type Attachment struct {
	Name string `json:"name" bson:"name"`
	Mime string `json:"mime" bson:"mime"`
	Data string `json:"data" bson:"data"`
}

type EmailID struct {
	ID string `json:"id"`
}

func (e *Emails) FromStorageModel(storageItems []storagemodel.Email) {
	for _, storageItem := range storageItems {
		var email Email
		email.FromStorageModel(&storageItem)
		*e = append(*e, email)
	}
}

type EmailsCount struct {
	Emails Emails `json:"emails"`
	Count  int64  `json:"count"`
}

func (e *EmailsCount) FromStorageModel(storageItems []storagemodel.Email, count int64) {

	e.Emails.FromStorageModel(storageItems)
	e.Count = count
}

func (e *Email) FromStorageModel(storageItem *storagemodel.Email) {
	e.ID = storageItem.ID
	e.Service = storageItem.Service
	e.To = storageItem.To
	e.Cc = storageItem.Cc
	e.Bcc = storageItem.Bcc
	e.Subject = storageItem.Subject
	e.Data = storageItem.Data
	e.HTML = storageItem.HTML
	e.Template = storageItem.Template
	e.Attachments.FromStorageModel(storageItem.Attachments)
	e.Log.FromStorageModel(storageItem.Log)
}

func (e *Email) ToStorageModel() *storagemodel.Email {
	return &storagemodel.Email{
		Service:     e.Service,
		To:          e.To,
		Cc:          e.Cc,
		Bcc:         e.Bcc,
		Subject:     e.Subject,
		Data:        e.Data,
		Template:    e.Template,
		Attachments: e.Attachments.ToStorageModel(),
	}
}

func (a *Attachments) FromStorageModel(storageItems []storagemodel.Attachment) {
	for _, storageItem := range storageItems {
		var attachment Attachment
		attachment.FromStorageModel(&storageItem)
		*a = append(*a, attachment)
	}
}

func (a Attachments) ToStorageModel() storagemodel.Attachments {

	var attachments storagemodel.Attachments

	for attachment := range a {
		attachments = append(attachments, a[attachment].ToStorageModel())
	}

	return attachments
}

func (a *Attachment) FromStorageModel(storageItem *storagemodel.Attachment) {
	a.Name = storageItem.Name
	a.Mime = storageItem.Mime
	a.Data = storageItem.Data
}

func (a *Attachment) ToStorageModel() storagemodel.Attachment {
	return storagemodel.Attachment{
		Name: a.Name,
		Mime: a.Mime,
		Data: a.Data,
	}
}
