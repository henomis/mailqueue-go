package log

import "github.com/henomis/mailqueue-go/pkg/email"

//Log struct
type Log struct {
	Service   string         `json:"appname" bson:"service"`
	Timestmap int64          `json:"timestamp" bson:"timestamp"`
	UUID      email.UniqueID `json:"uuid" bson:"uuid"`
	Status    int            `json:"status" bson:"status"`
	Error     string         `json:"error" bson:"error"`
}

//Logger interface
type Logger interface {
	Attach() error
	Detach() error
	Log(l *Log) error
	GetByUUID(uuid email.UniqueID) ([]*Log, error)
}
