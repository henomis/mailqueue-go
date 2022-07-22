package mongoemailqueue

import (
	"fmt"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/email"
	"github.com/henomis/mailqueue-go/internal/pkg/limiter"
	"github.com/henomis/mailqueue-go/internal/pkg/mongostorage"
	"github.com/pkg/errors"
)

type MongoEmailQueueOptions struct {
	Endpoint   string
	Database   string
	Collection string
	CappedSize uint64
	Timeout    time.Duration
}

type MongoEmailQueue struct {
	mongoQueueOptions *MongoEmailQueueOptions
	limiter           limiter.Limiter
	mongoStorage      *mongostorage.MongoStorage
}

func New(mongoQueueOptions *MongoEmailQueueOptions, limiter limiter.Limiter) (*MongoEmailQueue, error) {

	err := validateMongoQueueOptions(mongoQueueOptions)
	if err != nil {
		return nil, errors.Wrap(err, "invalid mongo queue options")
	}

	mongoStorage, err := mongostorage.New(
		mongoQueueOptions.Endpoint,
		mongoQueueOptions.Timeout,
		mongoQueueOptions.Database,
		mongoQueueOptions.Collection,
		mongoQueueOptions.CappedSize,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable create mongostorage")
	}

	err = mongoStorage.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "unable connect")
	}

	err = mongoStorage.CreateCappedCollection()
	if err != nil {
		return nil, errors.Wrap(err, "unable setup mongo capped collection")
	}

	mongoQueue := &MongoEmailQueue{
		mongoQueueOptions: mongoQueueOptions,
		limiter:           limiter,
		mongoStorage:      mongoStorage,
	}

	return mongoQueue, nil
}

func (q *MongoEmailQueue) Enqueue(email *email.Email) (string, error) {

	id, err := q.mongoStorage.InsertOne(email)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert data")
	}

	return id.(string), nil

}

func (q *MongoEmailQueue) Dequeue() (*email.Email, error) {

	filterQuery := mongostorage.Query(`{"sent": false}`)
	err := q.mongoStorage.WaitCappedCollectionCursor(filterQuery)
	if err != nil {
		return nil, errors.Wrap(err, "error waiting for capped collection cursor")
	}

	//waiting limiter
	<-q.limiter.Wait()

	var email email.Email
	err = q.mongoStorage.Decode(&email)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding data")
	}

	return &email, nil
}

func (q *MongoEmailQueue) SetProcessed(id string) error {

	filterQuery := mongostorage.Queryf(`{"_id": "%s"}`, id)
	updateQuery := mongostorage.Query(`{"$set": {"sent": true}}`)

	err := q.mongoStorage.Update(filterQuery, updateQuery)
	if err != nil {
		return errors.Wrap(err, "unable to update data")
	}

	return err
}

func (q *MongoEmailQueue) SetStatus(id string, status email.Status) error {

	filterQuery := mongostorage.Queryf(`{"_id": "%s"}`, id)
	updateQuery := mongostorage.Queryf(`{"$set": {"status": %d}}`, status)

	err := q.mongoStorage.Update(filterQuery, updateQuery)
	if err != nil {
		return errors.Wrap(err, "unable to update data")
	}

	return err
}

// ---------------
// Support methods
// ---------------

func validateMongoQueueOptions(mongoQueueOptions *MongoEmailQueueOptions) error {

	if len(mongoQueueOptions.Endpoint) == 0 {
		return fmt.Errorf("invalid endpoint")
	}

	if len(mongoQueueOptions.Database) == 0 {
		return fmt.Errorf("invalid database name")
	}

	if len(mongoQueueOptions.Collection) == 0 {
		return fmt.Errorf("invalid collection name")
	}

	if mongoQueueOptions.CappedSize == 0 {
		return fmt.Errorf("invalid capped size")
	}

	return nil
}
