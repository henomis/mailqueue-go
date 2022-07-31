package restmodel

import (
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/email"
)

type LogItems []LogItem
type LogItem struct {
	ID        string    `json:"id"`
	Service   string    `json:"appname"`
	Timestmap time.Time `json:"timestamp"`
	EmailID   string    `json:"email_id"`
	Status    int       `json:"status"`
	Error     string    `json:"error,omitempty"`
}

func (li *LogItems) FromStorage(storageItems []email.Log) {

	for _, storageItem := range storageItems {
		var logItem LogItem
		logItem.FromStorage(&storageItem)
		*li = append(*li, logItem)
	}
}

func (li *LogItem) FromStorage(storageItem *email.Log) {
	li.ID = storageItem.ID
	li.Service = storageItem.Service
	li.Timestmap = storageItem.Timestamp
	li.EmailID = storageItem.EmailID
	li.Status = storageItem.Status
	li.Error = storageItem.Error
}
