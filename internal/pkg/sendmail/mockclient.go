package sendmail

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/email"
)

type MockSMTPClient struct {
	attempts int
}

func New(attempts int) *MockSMTPClient {
	return &MockSMTPClient{
		attempts: attempts,
	}
}

func (c *MockSMTPClient) Send(e *email.Email) error {
	fmt.Printf("SENDING %+v\n", e)

	rand.Seed(time.Now().Unix())

	if rand.Intn(2) == 0 {
		return errors.New("SMTP ERROR")
	}

	return nil
}

func (c *MockSMTPClient) Attempts() int {
	return c.attempts
}
