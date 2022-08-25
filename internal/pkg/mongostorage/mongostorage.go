package mongostorage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoQuery bson.M
type MongoQuerya bson.A
type MongoFindOptions *options.FindOptions

type MongoStorage struct {
	timeout         time.Duration
	database        string
	collection      string
	cappedSize      uint64
	mongoClient     *mongo.Client
	mongoCollection *mongo.Collection
	mongoCursor     *mongo.Cursor
}

func New(
	endpoint string,
	timeout time.Duration,
	database string,
	collection string,
	cappedSize uint64,
) (*MongoStorage, error) {

	mongoClient, err := newMongoClient(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "unable create mongo client")
	}

	return &MongoStorage{
		timeout:     timeout,
		mongoClient: mongoClient,
		database:    database,
		collection:  collection,
		cappedSize:  cappedSize,
	}, nil
}

func (ms *MongoStorage) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	if err := ms.mongoClient.Connect(ctx); err != nil {
		return errors.Wrap(err, "unable to connect to mongodb")
	}

	if err := ms.mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		return errors.Wrap(err, "unable to ping mongodb")
	}

	return nil
}

func (ms *MongoStorage) CreateCollection() {
	db := ms.mongoClient.Database(ms.database)
	ms.mongoCollection = db.Collection(ms.collection)
}

func (ms *MongoStorage) CreateCappedCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	collectionExists := false
	db := ms.mongoClient.Database(ms.database)

	mongoCollections, err := db.ListCollectionNames(ctx, bson.D{}, nil)
	if err != nil {
		return errors.Wrap(err, "unable list collections names")
	}

	for _, mongoCollection := range mongoCollections {
		if mongoCollection == ms.collection {
			collectionExists = true
			break
		}
	}

	if !collectionExists {

		isTrue := true
		createCollectionOptions := &options.CreateCollectionOptions{
			Capped:      &isTrue,
			SizeInBytes: mongoCappedSizeFromUint64(ms.cappedSize),
		}
		err = db.CreateCollection(ctx, ms.collection, createCollectionOptions)
		if err != nil {
			return errors.Wrap(err, "unable create collection")
		}
	}

	ms.mongoCollection = db.Collection(ms.collection)

	return nil

}

func (ms *MongoStorage) InsertOne(data interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	insertOneResult, err := ms.mongoCollection.InsertOne(ctx, data)
	if err != nil {
		return nil, errors.Wrap(err, "unable insert data")
	}

	return insertOneResult.InsertedID, nil
}

func (ms *MongoStorage) FindOne(filterQuery MongoQuery, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	err := ms.mongoCollection.FindOne(ctx, filterQuery).Decode(data)
	if err != nil {
		return errors.Wrap(err, "unable find data")
	}

	return nil
}

func (ms *MongoStorage) ReplaceOrInsert(filterQuery MongoQuery, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	isTrue := true
	mongoReplaceOptions := &options.ReplaceOptions{
		Upsert: &isTrue,
	}
	_, err := ms.mongoCollection.ReplaceOne(ctx, filterQuery, data, mongoReplaceOptions)
	if err != nil {
		return errors.Wrap(err, "unable replace data")
	}

	return nil
}

func (ms *MongoStorage) DeleteOne(filterQuery MongoQuery) error {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	result, err := ms.mongoCollection.DeleteOne(ctx, filterQuery)
	if err != nil {
		return errors.Wrap(err, "unable to delete data")
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("document not found")
	}

	return nil
}

func (ms *MongoStorage) Decode(data interface{}) error {

	err := ms.mongoCursor.Decode(data)
	if err != nil {
		return errors.Wrap(err, "unable decode data")
	}

	return nil
}

func (ms *MongoStorage) DecodeAll(filterQuery MongoQuery, options MongoFindOptions, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	cursor, err := ms.mongoCollection.Find(ctx, filterQuery, options)
	if err != nil {
		return errors.Wrap(err, "unable to find data")
	}

	err = cursor.All(ctx, data)
	if err != nil {
		return errors.Wrap(err, "unable to fetch data")
	}

	return nil
}

func (ms *MongoStorage) Update(filterQuery MongoQuery, updateQuery interface{}) error {

	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	updateOneResult, err := ms.mongoCollection.UpdateOne(ctx, filterQuery, updateQuery)
	if err != nil {
		return errors.Wrap(err, "unable update data")
	}

	if updateOneResult.MatchedCount == 0 {
		return errors.New("no data matched")
	}

	return nil

}

func (ms *MongoStorage) Aggregate(pipeline MongoQuerya, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	cursor, err := ms.mongoCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return errors.Wrap(err, "unable aggregate data")
	}

	err = cursor.All(ctx, data)
	if err != nil {
		return errors.Wrap(err, "unable to fetch data")
	}

	return nil
}

func (ms *MongoStorage) Count(query MongoQuery) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	count, err := ms.mongoCollection.CountDocuments(ctx, query)
	if err != nil {
		return 0, errors.Wrap(err, "unable count data")
	}

	return count, nil
}

func (ms *MongoStorage) WaitCappedCollectionCursor(filterQuery MongoQuery) error {

	err := ms.setupTailableAwaitCursor(filterQuery)
	if err != nil {
		return errors.Wrap(err, "unable setup tailable await cursor")
	}

	err = ms.waitCursor(filterQuery)
	if err != nil {
		return errors.Wrap(err, "error waiting cursor")
	}

	return nil
}

func Query(query string) MongoQuery {

	var bsonMap bson.M

	err := json.Unmarshal([]byte(query), &bsonMap)
	if err != nil {
		return MongoQuery(bson.M{})
	}

	return MongoQuery(bsonMap)
}

func Queryf(query string, args ...interface{}) MongoQuery {
	return Query(fmt.Sprintf(query, args...))
}

func Querya(query string) MongoQuerya {

	var bsonArray bson.A

	err := json.Unmarshal([]byte(query), &bsonArray)
	if err != nil {
		return MongoQuerya(bson.A{})
	}

	return MongoQuerya(bsonArray)
}

func Queryaf(query string, args ...interface{}) MongoQuerya {
	return Querya(fmt.Sprintf(query, args...))
}

func SetSort(opts MongoFindOptions, query MongoQuery) MongoFindOptions {
	if opts == nil {
		opts = options.Find()
	}
	(*options.FindOptions)(opts).SetSort(query)

	return opts
}

func SetLimit(opts MongoFindOptions, limit int64) MongoFindOptions {
	if opts == nil {
		opts = options.Find()
	}
	(*options.FindOptions)(opts).SetLimit(limit)

	return opts
}

func SetSkip(opts MongoFindOptions, offset int64) MongoFindOptions {
	if opts == nil {
		opts = options.Find()
	}
	(*options.FindOptions)(opts).SetSkip(offset)

	return opts
}

func SetProjection(opts MongoFindOptions, fields []string) MongoFindOptions {

	if len(fields) == 0 {
		return opts
	}

	if opts == nil {
		opts = options.Find()
	}

	queryString := "{"
	for i, field := range fields {
		if i == 0 {
			queryString += fmt.Sprintf(`"%s": 1`, field)
		} else {
			queryString += fmt.Sprintf(`, "%s": 1`, field)
		}
	}
	queryString += "}"
	queryFilter := Query(queryString)

	(*options.FindOptions)(opts).SetProjection(queryFilter)

	return opts
}

func RandomID() string {
	return uuid.New().String()
}
