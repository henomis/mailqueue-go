package mongoemaillog

import (
	"fmt"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/email"
	"github.com/henomis/mailqueue-go/internal/pkg/mongostorage"
	"github.com/pkg/errors"
)

type MongoEmailLogOptions struct {
	Endpoint   string
	Database   string
	Collection string
	CappedSize uint64
	Timeout    time.Duration
}

type MongoEmailLog struct {
	mongoEmailLogOptions *MongoEmailLogOptions
	mongoStorage         *mongostorage.MongoStorage
}

func New(mongoEmailLogOptions *MongoEmailLogOptions) (*MongoEmailLog, error) {

	err := validateMongoEmailLogOptions(mongoEmailLogOptions)
	if err != nil {
		return nil, errors.Wrap(err, "invalid mongo email log options")
	}

	mongoStorage, err := mongostorage.New(
		mongoEmailLogOptions.Endpoint,
		mongoEmailLogOptions.Timeout,
		mongoEmailLogOptions.Database,
		mongoEmailLogOptions.Collection,
		mongoEmailLogOptions.CappedSize,
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
		return nil, errors.Wrap(err, "unable to create capped collection")
	}

	return &MongoEmailLog{
		mongoEmailLogOptions: mongoEmailLogOptions,
		mongoStorage:         mongoStorage,
	}, nil
}

func (mel *MongoEmailLog) Log(log *email.Log) (string, error) {

	log.ID = mongostorage.RandomID()
	log.Timestmap = time.Now().UTC()

	id, err := mel.mongoStorage.InsertOne(log)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert data")
	}

	return id.(string), nil
}

func (ml *MongoEmailLog) Items(emailID string) ([]email.Log, error) {

	var logItems []email.Log

	filterQuery := mongostorage.Queryf(`{"email_id": "%s"}`, emailID)
	err := ml.mongoStorage.DecodeAll(filterQuery, &logItems)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode data")
	}

	//TODO: this should check for invalid capped log
	//log items must contains the first entry status

	return logItems, err
}

// ---------------
// Support methods
// ---------------

func validateMongoEmailLogOptions(mongoEmailLogOptions *MongoEmailLogOptions) error {

	if len(mongoEmailLogOptions.Endpoint) == 0 {
		return fmt.Errorf("invalid endpoint")
	}

	if len(mongoEmailLogOptions.Database) == 0 {
		return fmt.Errorf("invalid database name")
	}

	if len(mongoEmailLogOptions.Collection) == 0 {
		return fmt.Errorf("invalid collection name")
	}

	return nil
}
