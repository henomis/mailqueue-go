package sendmail

import "github.com/henomis/mailqueue-go/internal/pkg/storagemodel"

//Client interface
type Client interface {
	Send(e *storagemodel.Email) error
	Attempts() int
}
