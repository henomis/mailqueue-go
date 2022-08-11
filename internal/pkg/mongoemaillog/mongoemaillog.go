package mongoemaillog

import (
	"strings"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/mongostorage"
	"github.com/henomis/mailqueue-go/internal/pkg/storagemodel"
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

func (mel *MongoEmailLog) Create(log *storagemodel.Log) (string, error) {

	log.ID = mongostorage.RandomID()
	log.Timestamp = time.Now().UTC()

	id, err := mel.mongoStorage.InsertOne(log)
	if err != nil {
		return "", errors.Wrap(err, "unable to insert data")
	}

	return id.(string), nil
}

func (ml *MongoEmailLog) Get(emailID string) ([]storagemodel.Log, error) {

	var logItems []storagemodel.Log
	var sortOptions mongostorage.MongoFindOptions

	filterQuery := mongostorage.Queryf(`{"email_id": "%s"}`, emailID)
	sortQuery := mongostorage.Queryf(`{"timestamp": 1}`)

	sortOptions = mongostorage.SetSort(sortOptions, sortQuery)

	err := ml.mongoStorage.DecodeAll(filterQuery, sortOptions, &logItems)
	if err != nil {
		return nil, errors.Wrap(err, "unable to decode data")
	}

	//TODO: this should check for invalid capped log
	//log items must contains the first entry status

	return logItems, err
}

func (ml *MongoEmailLog) GetAll(limit, skip int64, fields string) ([]storagemodel.Log, int64, error) {
	var storageLogs []storagemodel.Log

	findOptions := mongostorage.SetLimit(nil, limit)
	findOptions = mongostorage.SetSkip(findOptions, skip)
	if len(fields) > 0 {
		fieldsParts := strings.Split(fields, ",")
		findOptions = mongostorage.SetProjection(nil, fieldsParts)
	}

	count, err := ml.mongoStorage.Count(mongostorage.Query(""))
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable count templates")
	}

	err = ml.mongoStorage.DecodeAll(mongostorage.Query(""), findOptions, &storageLogs)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable find templates")
	}

	return storageLogs, count, nil
}
