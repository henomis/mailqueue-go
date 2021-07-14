package queue

import (
	"github.com/henomis/mailqueue-go/pkg/email"
)

const (
	//ErrLimitError limiter error
	ErrLimitError = "Limit reached"
	//ErrNotAttached client error
	ErrNotAttached = "Queue not attached"
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
