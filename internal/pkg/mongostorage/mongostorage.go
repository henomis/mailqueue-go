package mongostorage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoQuery bson.M

type MongoStorage struct {
	timeout         time.Duration
	database        string
	collection      string
	cappedSize      uint64
	filterQuery     MongoQuery
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
	filterQuery MongoQuery,
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
		filterQuery: filterQuery,
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

func (ms *MongoStorage) Insert(data interface{}) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	insertOneResult, err := ms.mongoCollection.InsertOne(ctx, data)
	if err != nil {
		return nil, errors.Wrap(err, "unable insert data")
	}

	return insertOneResult.InsertedID, nil
}

func (ms *MongoStorage) Decode(data interface{}) error {

	err := ms.mongoCursor.Decode(data)
	if err != nil {
		return errors.Wrap(err, "unable decode data")
	}

	return nil
}

func (ms *MongoStorage) DecodeAll(filterQuery MongoQuery, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	cursor, err := ms.mongoCollection.Find(ctx, filterQuery)
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

func (ms *MongoStorage) WaitCappedCollectionCursor() error {

	err := ms.setupTailableAwaitCursor(ms.filterQuery)
	if err != nil {
		return errors.Wrap(err, "unable setup tailable await cursor")
	}

	err = ms.waitCursor()
	if err != nil {
		return errors.Wrap(err, "error waiting cursor")
	}

	return nil
}

func (ms *MongoStorage) setupTailableAwaitCursor(query MongoQuery) error {

	if ms.mongoCursor != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	mongoFindOptions := options.Find().SetCursorType(options.TailableAwait).SetNoCursorTimeout(true)

	mongoCursor, err := ms.mongoCollection.Find(ctx, query, mongoFindOptions)
	if err != nil {
		return err
	}

	ms.mongoCursor = mongoCursor

	return nil
}

func (ms *MongoStorage) waitCursor() error {

	for {

		isNextDocumentAvailable := ms.mongoCursor.TryNext(context.Background())
		if isNextDocumentAvailable {
			//log.Println("mongo cursor has next document")
			break
		} else if ms.mongoCursor.ID() == 0 {
			//empty collection
			// log.Println("empty collection")
			time.Sleep(1 * time.Second)
			ms.setupTailableAwaitCursor(ms.filterQuery)
			continue
		} else if err := ms.mongoCursor.Err(); err != nil {
			ms.mongoCursor = nil
			return err
		} else {
			//log.Println("mongo cursor else sleep")
			// waiting element
			time.Sleep(1 * time.Second)
		}
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

// ---------------
// Support methods
// ---------------

func newMongoClient(endpoint string) (*mongo.Client, error) {
	mongoClientOptions := options.Client().ApplyURI(endpoint)
	err := mongoClientOptions.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "invalid mongodb endpoint")
	}

	return mongo.NewClient(mongoClientOptions)
}

func mongoCappedSizeFromUint64(cappedSize uint64) *int64 {
	cappedSizeAsInt64 := int64(cappedSize)
	return &cappedSizeAsInt64
}
