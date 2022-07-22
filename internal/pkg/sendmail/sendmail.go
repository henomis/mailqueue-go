package sendmail

import (
	"github.com/henomis/mailqueue-go/internal/pkg/email"
)

//Client interface
type Client interface {
	Send(e *email.Email) error
	Attempts() int
}

//Options for Sendmail Clients
type Options struct {
	Server   string
	Username string
	Password string

	From     string
	FromName string
	ReplyTo  string
	Attempts string
}
