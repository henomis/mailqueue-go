package sendmail

import (
	"github.com/henomis/mailqueue-go/internal/pkg/email"
)

//Client interface
type Client interface {
	Send(e *email.Email) error
	Attempts() int
}
