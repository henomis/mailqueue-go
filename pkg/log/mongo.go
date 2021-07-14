package log

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"github.com/henomis/mailqueue-go/pkg/email"
)

const (
	collectionLog = "log"
)

//MongoDBOptions for queue
type MongoDBOptions struct {
	Endpoint string
	Database string
	Timeout  time.Duration
}

//MongoDB queue implementation
type MongoDB struct {
	Options MongoDBOptions

	client *mongo.Client
	log    *mongo.Collection
}

//MongoLogEntry struct extends Log
type MongoLogEntry struct {
	ID primitive.ObjectID `json:"id" bson:"_id"`
	Log
}

func createContext(t time.Duration) (context.Context, context.CancelFunc) {
	if t == time.Duration(0) {
		return context.Background(), nil
	}

	return context.WithTimeout(context.Background(), t)
}

func callIfNotNil(c context.CancelFunc) func() {
	return func() {
		if c != nil {
			c()
		}
	}
}

//NewMongoDBLog creates a MongoDB queue instance
func NewMongoDBLog(opt MongoDBOptions) *MongoDB {

	return &MongoDB{
		Options: opt,
	}
}

//Attach memory queue
func (q *MongoDB) Attach() error {

	ctx, cancel := createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()

	opts := options.Client().ApplyURI(q.Options.Endpoint)
	err := opts.Validate()
	if err != nil {
		return err
	}

	q.client, err = mongo.NewClient(opts)
	if err != nil {
		return err
	}

	ctx, cancel = createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()
	if err := q.client.Connect(ctx); err != nil {
		return err
	}

	ctx, cancel = createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()
	if err := q.client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	db := q.client.Database(q.Options.Database)

	/*** LOG *******/

	q.log = db.Collection(collectionLog)
	if q.log == nil {
		return errors.New("no collection")
	}

	return nil

}

//Detach implementation in memory
func (q *MongoDB) Detach() error {

	ctx, cancel := createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()
	err := q.log.Database().Client().Disconnect(ctx)
	if err != nil {
		return err
	}

	q.client = nil
	q.log = nil

	return nil
}

//Log implementation in memory
func (q *MongoDB) Log(l *Log) error {

	l.Timestmap = time.Now().UTC().UnixNano()

	ctx, cancel := createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()

	if l == nil {
		return errors.New("cannot log nil entry")
	}

	doc, err := bson.Marshal(l)
	if err != nil {
		return err
	}

	_, err = q.log.InsertOne(ctx, doc)
	if err != nil {
		return err
	}

	return nil
}

//GetByUUID func
func (q *MongoDB) GetByUUID(uuid email.UniqueID) ([]*Log, error) {

	var logEntries []*Log

	ctx, cancel := createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()

	opt := options.Find().SetSort(bson.M{"timestamp": 1})

	filter := bson.M{"uuid": uuid}
	cur, err := q.log.Find(ctx, filter, opt)
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	ctx, cancel = createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()

	for cur.Next(ctx) {
		e := &Log{}
		err := cur.Decode(e)
		if err != nil {
			return nil, err
		}
		logEntries = append(logEntries, e)

	}

	return logEntries, nil
}
