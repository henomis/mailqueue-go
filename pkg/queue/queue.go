package queue

import (
	"github.com/henomis/mailqueue-go/pkg/email"
)

const (
	//ErrLimitError limiter error
	ErrLimitError = "limit reached"
	//ErrNotAttached client error
	ErrNotAttached = "queue not attached"
)

//Queue implementation
type Queue interface {
	Attach() error
	Detach() error
	Enqueue(*email.Email) (email.UniqueID, error)
	Dequeue() (*email.Email, error)
	SetStatus(*email.Email, email.Status) error
	Commit(*email.Email) error
	GetByUUID(uuid email.UniqueID) (*email.Email, error)
}
