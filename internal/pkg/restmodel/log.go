package restmodel

import (
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/storagemodel"
)

type Logs []Log

func (l *Logs) FromStorageModel(storageItems []storagemodel.Log) {
	for _, storageItem := range storageItems {
		var log Log
		log.FromStorageModel(&storageItem)
		*l = append(*l, log)
	}
}

type Log struct {
	ID        string    `json:"id"`
	Service   string    `json:"service"`
	Timestmap time.Time `json:"timestamp"`
	EmailID   string    `json:"email_id"`
	Status    int       `json:"status"`
	Error     string    `json:"error,omitempty"`
}

func (li *Log) FromStorageModel(storageItem *storagemodel.Log) {
	li.ID = storageItem.ID
	li.Service = storageItem.Service
	li.Timestmap = storageItem.Timestamp
	li.EmailID = storageItem.EmailID
	li.Status = storageItem.Status
	li.Error = storageItem.Error
}

type LogsCount struct {
	Logs  Logs  `json:"logs"`
	Count int64 `json:"count"`
}

func (l *LogsCount) FromStorageModel(storageItems []storagemodel.Log, count int64) {

	l.Logs.FromStorageModel(storageItems)
	l.Count = count
}
