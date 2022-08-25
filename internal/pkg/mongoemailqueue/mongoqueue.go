package mongoemailqueue

import (
	"bytes"
	"io"
	"strings"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/mongostorage"
	"github.com/henomis/mailqueue-go/internal/pkg/storagemodel"
	"github.com/pkg/errors"
)

type Limiter interface {
	Wait() chan struct{}
}

type EmailTemplate interface {
	Execute(inputDataReader io.Reader, outputDataWriter io.Writer, templateID string) error
}

type MongoEmailQueueOptions struct {
	Endpoint   string
	Database   string
	Collection string
	CappedSize uint64
	Timeout    time.Duration
}

type MongoEmailQueue struct {
	mongoQueueOptions *MongoEmailQueueOptions
	limiter           Limiter
	mongoStorage      *mongostorage.MongoStorage
	template          EmailTemplate
}

func New(mongoQueueOptions *MongoEmailQueueOptions, limiter Limiter, mongoTemplate EmailTemplate) (*MongoEmailQueue, error) {

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
		template:          mongoTemplate,
	}

	return mongoQueue, nil
}

func (q *MongoEmailQueue) Enqueue(email *storagemodel.Email) (string, error) {

	email.ID = mongostorage.RandomID()

	if len(email.Template) > 0 && q.template != nil {

		var buffer bytes.Buffer
		err := q.template.Execute(strings.NewReader(email.Data), io.Writer(&buffer), email.Template)
		if err != nil {
			return "", errors.Wrap(err, "unable to render email")
		}

		email.HTML = buffer.String()
	}

	id, err := q.mongoStorage.InsertOne(email)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert data")
	}

	return id.(string), nil

}

func (q *MongoEmailQueue) Dequeue() (*storagemodel.Email, error) {

	filterQuery := mongostorage.Query(`{"processed": false}`)
	err := q.mongoStorage.WaitCappedCollectionCursor(filterQuery)
	if err != nil {
		return nil, errors.Wrap(err, "error waiting for capped collection cursor")
	}

	//waiting limiter
	<-q.limiter.Wait()

	var email storagemodel.Email
	err = q.mongoStorage.Decode(&email)
	if err != nil {
		return nil, errors.Wrap(err, "error decoding data")
	}

	return &email, nil
}

func (q *MongoEmailQueue) SetProcessed(id string) error {

	filterQuery := mongostorage.Queryf(`{"_id": "%s"}`, id)
	updateQuery := mongostorage.Query(`{"$set": {"processed": true}}`)

	err := q.mongoStorage.Update(filterQuery, updateQuery)
	if err != nil {
		return errors.Wrap(err, "unable to update data")
	}

	return err
}

func (q *MongoEmailQueue) SetStatus(id string, status storagemodel.Status) error {

	filterQuery := mongostorage.Queryf(`{"_id": "%s"}`, id)
	updateQuery := mongostorage.Queryf(`{"$set": {"status": %d}}`, status)

	if status == storagemodel.StatusRead {
		filterQuery = mongostorage.Queryf(`{"_id": "%s", "status": {"$in": [%d,%d]}}`,
			id, storagemodel.StatusSent, storagemodel.StatusRead)
	}

	err := q.mongoStorage.Update(filterQuery, updateQuery)
	if err != nil {
		return errors.Wrap(err, "unable to update data")
	}

	return err
}

func (q *MongoEmailQueue) Get(id string) (*storagemodel.Email, error) {
	var mongoEmail storagemodel.Email

	filterQuery := mongostorage.Queryf(`{"_id": "%s"}`, id)
	err := q.mongoStorage.FindOne(filterQuery, &mongoEmail)
	if err != nil {
		return nil, errors.Wrap(err, "unable find email")
	}

	return &mongoEmail, nil
}

func (q *MongoEmailQueue) GetAll(limit, skip int64, fields string) ([]storagemodel.Email, int64, error) {
	var storageEmails []storagemodel.Email

	findOptions := mongostorage.SetLimit(nil, limit)
	findOptions = mongostorage.SetSkip(findOptions, skip)
	if len(fields) > 0 {
		fieldsParts := strings.Split(fields, ",")
		findOptions = mongostorage.SetProjection(nil, fieldsParts)
	}

	count, err := q.mongoStorage.Count(mongostorage.Query(""))
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable count templates")
	}

	err = q.mongoStorage.DecodeAll(mongostorage.Query(""), findOptions, &storageEmails)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable find templates")
	}

	return storageEmails, count, nil
}

func (q *MongoEmailQueue) GetAllWithLogs(limit, skip int64) ([]storagemodel.Email, int64, error) {
	var storageEmails []storagemodel.Email

	aggregationQuery := mongostorage.Queryaf(
		`[
			{
			  "$lookup": {
				"from": "log", 
				"localField": "_id", 
				"foreignField": "email_id", 
				"as": "log"
			  }
			}, {
			  "$limit": %d
			}, {
			  "$skip": %d
			}
		  ]`,
		limit,
		skip,
	)
	err := q.mongoStorage.Aggregate(aggregationQuery, &storageEmails)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable find emails")
	}

	count, err := q.mongoStorage.Count(mongostorage.Query(""))
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable count emails")
	}

	return storageEmails, count, nil
}
