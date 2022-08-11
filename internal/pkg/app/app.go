package app

import (
	"github.com/henomis/mailqueue-go/internal/pkg/storagemodel"
)

type App interface {
	Run() error
}

type EmailQueue interface {
	Enqueue(email *storagemodel.Email) (string, error)
	Dequeue() (*storagemodel.Email, error)
	SetProcessed(id string) error
	SetStatus(id string, status storagemodel.Status) error
	Get(id string) (*storagemodel.Email, error)
	GetAll(limit, skip int64, fields string) ([]storagemodel.Email, int64, error)
}

type EmailLog interface {
	Create(log *storagemodel.Log) (string, error)
	Get(emailID string) ([]storagemodel.Log, error)
	GetAll(limit, skip int64, fields string) ([]storagemodel.Log, int64, error)
}

type EmailTemplate interface {
	Create(mongoTemplate *storagemodel.Template) (string, error)
	Get(id string) (*storagemodel.Template, error)
	GetAll(limit, skip int64, fields string) ([]storagemodel.Template, int64, error)
	Update(id string, mongoTemplate *storagemodel.Template) error
	Delete(id string) error
}
