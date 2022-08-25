package storagemodel

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
	ID          string      `bson:"_id"`
	Service     string      `bson:"service"`
	To          string      `bson:"to"`
	Cc          string      `bson:"cc"`
	Bcc         string      `bson:"bcc"`
	Subject     string      `bson:"subject"`
	HTML        string      `bson:"html"`
	Data        string      `bson:"data"`
	Template    string      `bson:"template"`
	Attachments Attachments `bson:"attachments"`
	Processed   bool        `bson:"processed"`
	Status      uint64      `bson:"status"`
	Log         []Log       `bson:"log"`
}

type Attachments []Attachment

type Attachment struct {
	Name string `json:"name" bson:"name"`
	Mime string `json:"mime" bson:"mime"`
	Data string `json:"data" bson:"data"`
}
