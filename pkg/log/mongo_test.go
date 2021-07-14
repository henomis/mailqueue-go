package log

import (
	"testing"
	"time"

	"github.com/henomis/mailqueue-go/pkg/email"
)

func TestLog(t *testing.T) {

	l := NewMongoDBLog(MongoDBOptions{Endpoint: "mongodb://admin:pass@localhost:27017/admin?authSource=admin", Database: "test", Timeout: 10 * time.Second})

	err := l.Attach()
	if err != nil {
		t.Errorf("Expected nil got %s", err.Error())
	}

	t.Run("test Insert log", func(t *testing.T) {
		e := &Log{
			Service: "service",
			Status:  email.StatusErrorSending,
			Error:   "Errore",
			UUID:    "1234",
		}

		err := l.Log(e)
		if err != nil {
			t.Errorf("Expected nil got %s", err.Error())
		}

	})
}
