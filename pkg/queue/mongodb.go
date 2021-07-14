package queue

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/henomis/mailqueue-go/pkg/email"
	"github.com/henomis/mailqueue-go/pkg/limiter"
	"github.com/henomis/mailqueue-go/pkg/render"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	collectionQueue = "queue"
)

//MongoDBOptions for queue
type MongoDBOptions struct {
	Endpoint   string
	Database   string
	CappedSize int64
	Timeout    time.Duration
}

//MongoDB queue implementation
type MongoDB struct {
	Options  MongoDBOptions
	Limiter  limiter.Limiter
	Template render.Render

	client      *mongo.Client
	queue       *mongo.Collection
	queueCursor *mongo.Cursor
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

//NewMongoDBQueue creates a MongoDB queue instance
func NewMongoDBQueue(opt MongoDBOptions, lim limiter.Limiter, tmpl render.Render) *MongoDB {

	return &MongoDB{
		Options:  opt,
		Limiter:  lim,
		Template: tmpl,
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

	/*** QUEUE *******/

	ctx, cancel = createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()
	exists := false
	colls, err := db.ListCollectionNames(ctx, bson.D{}, nil)
	if err != nil {
		return err
	}

	for _, v := range colls {
		if v == collectionQueue {
			exists = true
			break
		}
	}

	if !exists {

		//db.Collection(collectionQueue).Drop(ctx)

		ttrue := true
		cc := &options.CreateCollectionOptions{
			Capped:      &ttrue,
			SizeInBytes: &(q.Options.CappedSize),
		}

		ctx, cancel = createContext(q.Options.Timeout)
		defer callIfNotNil(cancel)()
		err = db.CreateCollection(ctx, collectionQueue, cc)
		if err != nil {
			return err
		}
	}

	q.queue = db.Collection(collectionQueue)

	return nil

}

//Detach implementation in memory
func (q *MongoDB) Detach() error {

	ctx, cancel := createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()
	err := q.queue.Database().Client().Disconnect(ctx)
	if err != nil {
		return err
	}

	q.client = nil
	q.queue = nil

	return nil
}

//Enqueue implementation in memory
func (q *MongoDB) Enqueue(e *email.Email) (email.UniqueID, error) {

	if q.client == nil {
		return "", errors.New(ErrNotAttached)
	}

	//HTML render
	bufferReader := bytes.NewBuffer([]byte(e.Data))
	buff := []byte{}
	bufferWriter := bytes.NewBuffer(buff)

	err := q.Template.Execute(bufferReader, bufferWriter, render.Key(e.Template))
	if err != nil {
		return "", err
	}

	e.Data = string(bufferWriter.Bytes())

	//assign UUID
	uuid := uuid.New()
	e.UUID = email.UniqueID(uuid.String())

	doc, err := bson.Marshal(e)
	if err != nil {
		return "", err
	}

	ctx, cancel := createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()
	_, err = q.queue.InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}

	return e.UUID, nil

}

//Dequeue implementation in mongoDB. This may block
func (q *MongoDB) Dequeue() (*email.Email, error) {

	if q.client == nil {
		return nil, errors.New(ErrNotAttached)
	}

	filter := bson.M{"sent": false}

	if q.queueCursor == nil {
		ctx, cancel := createContext(q.Options.Timeout)
		defer callIfNotNil(cancel)()

		opts := options.Find().SetCursorType(options.TailableAwait).SetNoCursorTimeout(true)
		cursor, err := q.queue.Find(ctx, filter, opts)
		if err != nil {
			return nil, err
		}
		q.queueCursor = cursor
	}

	//check empty collection
	for {

		ctx, cancel := createContext(q.Options.Timeout)
		defer callIfNotNil(cancel)()

		res := q.queueCursor.TryNext(ctx)
		if res == true {
			break
		} else if q.queueCursor.ID() == 0 {
			//empty collection
			time.Sleep(1 * time.Second)

			opts := options.Find().SetCursorType(options.TailableAwait).SetNoCursorTimeout(true)
			q.queueCursor, _ = q.queue.Find(context.Background(), filter, opts)

			continue
		} else if err := q.queueCursor.Err(); err != nil {
			return nil, q.queueCursor.Err()
		} else {
			// waiting element
			time.Sleep(1 * time.Second)
		}

	}

	//waiting limiter
	for {
		if q.Limiter.Allow() {
			break
		}
		//waiting limiter
		time.Sleep(1 * time.Second)
	}

	email := &email.Email{}
	if err := q.queueCursor.Decode(email); err != nil {
		return nil, err
	}

	return email, nil

}

//Commit mongodb implementation
func (q *MongoDB) Commit(e *email.Email) error {
	ctx, cancel := createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()

	update := bson.M{"$set": bson.M{"sent": true, "status": email.StatusSent}}
	_, err := q.queue.UpdateOne(ctx, bson.M{"uuid": e.UUID}, update)
	if err != nil {
		return err
	}
	return nil
}

//SetStatus mongodb implementation
func (q *MongoDB) SetStatus(e *email.Email, status email.Status) error {
	ctx, cancel := createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()

	update := bson.M{"$set": bson.M{"status": status}}
	_, err := q.queue.UpdateOne(ctx, bson.M{"uuid": e.UUID}, update)
	if err != nil {
		return err
	}
	return nil
}

//GetByUUID mongodb implementation
func (q *MongoDB) GetByUUID(uuid email.UniqueID) (*email.Email, error) {

	ctx, cancel := createContext(q.Options.Timeout)
	defer callIfNotNil(cancel)()

	filter := bson.M{"uuid": uuid}
	e := &email.Email{}

	err := q.queue.FindOne(ctx, filter).Decode(e)
	if err != nil {
		return nil, err
	}

	return e, nil
}
