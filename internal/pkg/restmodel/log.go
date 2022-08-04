package restmodel

import (
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/storagemodel"
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

func (li *LogItems) FromStorageModel(storageItems []storagemodel.Log) {

	for _, storageItem := range storageItems {
		var logItem LogItem
		logItem.FromStorageModel(&storageItem)
		*li = append(*li, logItem)
	}
}

func (li *LogItem) FromStorageModel(storageItem *storagemodel.Log) {
	li.ID = storageItem.ID
	li.Service = storageItem.Service
	li.Timestmap = storageItem.Timestamp
	li.EmailID = storageItem.EmailID
	li.Status = storageItem.Status
	li.Error = storageItem.Error
}
