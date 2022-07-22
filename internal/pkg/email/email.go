package email

//Status tipo dello stato dell'email
type Status int

const (
	StatusUnknown = iota
	//StatusQueued email inserita in coda
	StatusQueued
	//StatusDequeued email prelevata dalla coda grazie al limitatore
	StatusDequeued
	//StatusSending email in carico al client SMTP
	StatusSending
	//StatusSent email correttamente inviata
	StatusSent
	//StatusRead email letta
	StatusRead
	//StatusErrorSending errore SMTP client
	StatusErrorSending
	//StatusErrorCanceled email rimossa dall'invio
	StatusErrorCanceled
)

//Email represent email structure
type Email struct {
	ID          string       `json:"uuid" bson:"_id"`
	Service     string       `json:"service" bson:"service"`
	From        string       `json:"from" bson:"from"`
	FromName    string       `json:"fromname" bson:"fromname"`
	ReplyTo     string       `json:"replyto" bson:"replyto"`
	To          string       `json:"to" bson:"to"`
	Cc          string       `json:"cc" bson:"cc"`
	Bcc         string       `json:"bcc" bson:"bcc"`
	Subject     string       `json:"subject" bson:"subject"`
	HTML        string       `json:"html" bson:"html"`
	Data        string       `json:"data" bson:"data"`
	Template    string       `json:"template" bson:"template"`
	Attachments []Attachment `json:"attachments" bson:"attachments"`
	Sent        bool         `json:"sent" bson:"sent"`
	Status      uint64       `json:"status" bson:"status"`
}

//Attachment struct
type Attachment struct {
	Name string `json:"name" bson:"name"`
	Mime string `json:"mime" bson:"mime"`
	Data string `json:"data" bson:"data"`
}
