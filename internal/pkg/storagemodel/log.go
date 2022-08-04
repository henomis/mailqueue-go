package storagemodel

import "time"

type Log struct {
	ID        string    `json:"id" bson:"_id"`
	Service   string    `json:"appname" bson:"service"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	EmailID   string    `json:"email_id" bson:"email_id"`
	Status    int       `json:"status" bson:"status"`
	Error     string    `json:"error" bson:"error,omitempty"`
}
