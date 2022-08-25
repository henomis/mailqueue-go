package storagemodel

import "time"

type Log struct {
	ID        string    `bson:"_id"`
	Service   string    `bson:"service"`
	Timestamp time.Time `bson:"timestamp"`
	EmailID   string    `bson:"email_id"`
	Status    int       `bson:"status"`
	Error     string    `bson:"error,omitempty"`
}
