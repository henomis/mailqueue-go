package restmodel

import "github.com/henomis/mailqueue-go/internal/pkg/email"

type Email struct {
	Service     string      `json:"service"`
	To          string      `json:"to"`
	Cc          string      `json:"cc"`
	Bcc         string      `json:"bcc"`
	Subject     string      `json:"subject"`
	Data        string      `json:"data"`
	Template    string      `json:"template"`
	Attachments Attachments `json:"attachments"`
}

type Attachments []Attachment

type Attachment struct {
	Name string `json:"name" bson:"name"`
	Mime string `json:"mime" bson:"mime"`
	Data string `json:"data" bson:"data"`
}

func (a Attachments) ToStorageAttachment() email.Attachments {

	var attachments email.Attachments

	for attachment := range a {
		attachments = append(attachments, a[attachment].ToStorageAttachment())
	}

	return attachments
}

func (a *Attachment) ToStorageAttachment() email.Attachment {
	return email.Attachment{
		Name: a.Name,
		Mime: a.Mime,
		Data: a.Data,
	}
}

func (e *Email) ToStorageEmail() *email.Email {
	return &email.Email{
		Service:     e.Service,
		To:          e.To,
		Cc:          e.Cc,
		Bcc:         e.Bcc,
		Subject:     e.Subject,
		Data:        e.Data,
		Template:    e.Template,
		Attachments: e.Attachments.ToStorageAttachment(),
	}
}
