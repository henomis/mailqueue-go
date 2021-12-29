package queue

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/henomis/mailqueue-go/pkg/email"
	"github.com/henomis/mailqueue-go/pkg/limiter"
	mongorender "github.com/henomis/mailqueue-go/pkg/render/mongo"
)

var uuidTmp email.UniqueID

func TestMongoDB(t *testing.T) {

	uuidTmp = "xxx"
	q := MongoDB{
		Limiter: &limiter.DefaultLimiter{
			Allowed:  10,
			Interval: 1 * time.Minute,
		},
		Options: MongoDBOptions{
			CappedSize: 1000000,
			Database:   "test",
			Endpoint:   "mongodb://admin:pass@localhost:27017/admin?authSource=admin", //os.Getenv("MONGO_ENDPOINT"), //
			Timeout:    0,
		},
		/*Template: &render.FileRender{
			Path: "/tmp/",
		},*/
		Template: &mongorender.MongoRender{
			MongoDatabase: "test",
			MongoEndpoint: "mongodb://admin:pass@localhost:27017/admin?authSource=admin", //os.Getenv("MONGO_ENDPOINT"), //
			MongoTimeout:  10 * time.Second,
		},
	}

	t.Run("test attach", func(t *testing.T) {

		t.Helper()

		err := q.Attach()
		if err != nil {
			t.Errorf(err.Error())
		}

	})

	t.Run("test enqueue", func(t *testing.T) {

		t.Helper()
		var err error

		html := "<html><body>Hello, {{.world}}</body></html>"
		tmpl := []byte(`{"world":"world!"}`)

		m := make(map[string]interface{})
		if err := json.Unmarshal(tmpl, &m); err != nil {
			t.Errorf(err.Error())
		}

		uuidTmp, err = q.Enqueue(&email.Email{From: "pippo", Data: html, Template: "test"})
		if err != nil {
			t.Errorf(err.Error())
		}

	})

	t.Run("test dequeue", func(t *testing.T) {

		t.Helper()

		e, err := q.Dequeue()
		if err != nil {
			t.Errorf(err.Error())
		}

		if e.UUID != uuidTmp {
			t.Errorf("Expected %q got %q", uuidTmp, e.UUID)
		}

	})

	t.Run("test commit", func(t *testing.T) {

		t.Helper()
		e := &email.Email{
			UUID: uuidTmp,
		}

		err := q.Commit(e)
		if err != nil {
			t.Errorf(err.Error())
		}

	})

	t.Run("test setstatus", func(t *testing.T) {
		t.Helper()

		e := &email.Email{
			UUID: uuidTmp,
		}

		err := q.SetStatus(e, email.StatusDequeued)
		if err != nil {
			t.Errorf(err.Error())
		}
	})

	t.Run("test detach", func(t *testing.T) {

		t.Helper()

		err := q.Detach()
		if err != nil {
			t.Errorf(err.Error())
		}

	})
}
