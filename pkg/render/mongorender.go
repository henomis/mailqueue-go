package render

import (
	"context"
	"io"
	"io/ioutil"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const (
	collectionTemplate = "template"
)

// MongoTemplate struct
type MongoTemplate struct {
	Name        string `json:"name" bson:"name"`
	Tmpl        string `json:"tmpl" bson:"tmpl"`
	Placeholder string `json:"ph" bson:"ph"`
}

//MongoRender implementation with file
type MongoRender struct {
	Endpoint string
	Database string
	Timeout  time.Duration

	client     *mongo.Client
	collection *mongo.Collection
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

//NewMongoRender create a new instance
func NewMongoRender(tmo time.Duration, endpoint, db string) *MongoRender {
	return &MongoRender{
		Endpoint: endpoint,
		Database: db,
		Timeout:  tmo,
	}
}

//Set implemetation
func (mr *MongoRender) Set(k Key, v Value) error {

	if mr.client == nil {
		err := mr.attach()
		if err != nil {
			return err
		}
	}

	ttrue := true
	filter := bson.M{"name": k}
	upsert := &options.ReplaceOptions{
		Upsert: &ttrue,
	}

	e := &MongoTemplate{
		Name:        string(k),
		Placeholder: "ph",
		Tmpl:        string(v),
	}

	doc, err := bson.Marshal(e)
	if err != nil {
		return err
	}

	ctx, cancel := createContext(mr.Timeout)
	defer callIfNotNil(cancel)()
	_, err = mr.collection.ReplaceOne(ctx, filter, doc, upsert)
	if err != nil {
		return err
	}

	return nil
}

//Get implemetation
func (mr *MongoRender) Get(k Key) (Value, error) {

	if mr.client == nil {
		err := mr.attach()
		if err != nil {
			return nil, err
		}
	}

	filter := bson.M{"name": k}
	e := &MongoTemplate{}

	ctx, cancel := createContext(mr.Timeout)
	defer callIfNotNil(cancel)()
	err := mr.collection.FindOne(ctx, filter).Decode(e)
	if err != nil {
		return nil, err
	}

	return Value(e.Tmpl), nil

}

//Execute implemetation
func (mr *MongoRender) Execute(r io.Reader, w io.Writer, k Key) error {

	if mr.client == nil {
		err := mr.attach()
		if err != nil {
			return err
		}
	}

	v, err := mr.Get(k)
	if err != nil {
		return err
	}

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	j, err := createJSON(data)
	if err != nil {
		return err
	}

	err = merge(v, w, j)
	if err != nil {
		return err
	}

	return nil
}

func (mr *MongoRender) attach() error {

	opts := options.Client().ApplyURI(mr.Endpoint)
	err := opts.Validate()
	if err != nil {
		return err
	}

	mr.client, err = mongo.NewClient(opts)
	if err != nil {
		return err
	}

	ctx, cancel := createContext(mr.Timeout)
	defer callIfNotNil(cancel)()
	if err := mr.client.Connect(ctx); err != nil {
		return err
	}

	ctx, cancel = createContext(mr.Timeout)
	defer callIfNotNil(cancel)()
	if err := mr.client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	db := mr.client.Database(mr.Database)

	exists := false
	ctx, cancel = createContext(mr.Timeout)
	defer callIfNotNil(cancel)()
	colls, err := db.ListCollectionNames(ctx, bson.D{}, nil)
	if err != nil {
		return err
	}

	for _, v := range colls {
		if v == collectionTemplate {
			exists = true
			break
		}
	}

	if !exists {
		ctx, cancel := createContext(mr.Timeout)
		defer callIfNotNil(cancel)()
		err = db.CreateCollection(ctx, collectionTemplate, nil)
		if err != nil {
			return err
		}
	}

	mr.collection = db.Collection(collectionTemplate)

	return nil

}
