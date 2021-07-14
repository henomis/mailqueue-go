package email

//Status tipo dello stato dell'email
type Status int

const (
	//StatusQueued email inserita in coda
	StatusQueued = iota
	//StatusDequeued email prelevata dalla coda grazie al limitatore
	StatusDequeued
	//StatusSending email in carico al client SMTP
	StatusSending
	//StatusErrorSending errore SMTP client
	StatusErrorSending
	//StatusErrorCanceled email rimossa dall'invio
	StatusErrorCanceled
	//StatusSent email correttamente inviata
	StatusSent
	//StatusRead email letta
	StatusRead
)

//Attachment struct
type Attachment struct {
	Name string `json:"name"`
	Mime string `json:"mime"`
	Data string `json:"data"`
}

//UniqueID alias string
type UniqueID string

//Email represent email structure
type Email struct {
	UUID     UniqueID `json:"uuid"`
	Service  string   `json:"service"`
	From     string   `json:"from"`
	FromName string   `json:"fromname"`
	ReplyTo  string   `json:"replyto"`
	To       string   `json:"to"`
	Cc       string   `json:"cc"`
	Bcc      string   `json:"bcc"`
	Subject  string   `json:"subject"`
	//HTML        string       `json:"html"`
	Data        string       `json:"data" bson:"data"`
	Template    string       `json:"template" bson:"template"`
	Attachments []Attachment `json:"attachments"`
	Sent        bool         `json:"sent"`
	Status      int          `json:"status"`
}
