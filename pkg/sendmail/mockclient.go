package sendmail

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/henomis/mailqueue-go/pkg/email"
)

//MockSMTPClient sends email using MailYak package
type MockSMTPClient struct {
	options *Options
}

//NewMockSMTPClient returns a new MailYak smtp client
func NewMockSMTPClient(options *Options) *MockSMTPClient {

	return &MockSMTPClient{
		options: options,
	}

}

//Send implements sending email
func (c *MockSMTPClient) Send(e *email.Email) error {
	fmt.Printf("SENDING %+v\n", e)

	rand.Seed(time.Now().Unix())

	if rand.Intn(2) == 0 {
		return errors.New("SMTP ERROR")
	}

	return nil
}

//Attempts return attempts
func (c *MockSMTPClient) Attempts() int {
	a, _ := strconv.Atoi(c.options.Attempts)
	return a
}
