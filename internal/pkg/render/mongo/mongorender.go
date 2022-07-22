package mongorender

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/render"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collectionTemplate = "template"
)

// MongoTemplate struct
type MongoTemplate struct {
	Name     string `json:"name" bson:"name"`
	Template string `json:"template" bson:"template"`
}

//MongoRender implementation with file
type MongoRender struct {
	MongoEndpoint string
	MongoDatabase string
	MongoTimeout  time.Duration

	mongoClient     *mongo.Client
	mongoCollection *mongo.Collection
}

//NewMongoRender create a new instance
func NewMongoRender(mongoTimeout time.Duration, mongoEndpoint, mongoDatabase string) (*MongoRender, error) {

	mongoRender := &MongoRender{
		MongoEndpoint: mongoEndpoint,
		MongoDatabase: mongoDatabase,
		MongoTimeout:  mongoTimeout,
	}

	err := createMongoConnection(mongoRender)
	if err != nil {
		return nil, err
	}

	mongoRender.selectDatabaseAndCollection()

	return mongoRender, nil
}

//Set implemetation
func (mr *MongoRender) Set(mongoKey string, mongoValue interface{}) error {

	isTrue := true
	mongoFilter := bson.M{"name": mongoKey}
	mongoReplaceOptions := &options.ReplaceOptions{
		Upsert: &isTrue,
	}

	mongoValueAsString, ok := mongoValue.(string)
	if !ok {
		return fmt.Errorf("invalid value")
	}

	mongoTemplate := &MongoTemplate{
		Name:     mongoKey,
		Template: mongoValueAsString,
	}

	mongoTemplateRaw, err := bson.Marshal(mongoTemplate)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), mr.MongoTimeout)
	defer cancel()

	_, err = mr.mongoCollection.ReplaceOne(ctx, mongoFilter, mongoTemplateRaw, mongoReplaceOptions)
	if err != nil {
		return err
	}

	return nil
}

//Get implemetation
func (mr *MongoRender) Get(k string) (interface{}, error) {

	mongoFilter := bson.M{"name": k}
	mongoTemplate := &MongoTemplate{}

	ctx, cancel := context.WithTimeout(context.Background(), mr.MongoTimeout)
	defer cancel()

	err := mr.mongoCollection.FindOne(ctx, mongoFilter).Decode(mongoTemplate)
	if err != nil {
		return nil, err
	}

	return mongoTemplate.Template, nil

}

//Execute implemetation
func (mr *MongoRender) Execute(inputDataReader io.Reader, outputDataWriter io.Writer, mongoKey string) error {

	mongoTemplateBody, err := mr.Get(mongoKey)
	if err != nil {
		return err
	}

	mongoTemplateBodyAsString, ok := mongoTemplateBody.(string)
	if !ok {
		return fmt.Errorf("invalid template body")
	}

	rawDataFromReader, err := ioutil.ReadAll(inputDataReader)
	if err != nil {
		return err
	}

	templateDataObject, err := render.CreateTemplateDataObject(rawDataFromReader)
	if err != nil {
		return err
	}

	err = render.Merge(mongoTemplateBodyAsString, templateDataObject, outputDataWriter)
	if err != nil {
		return err
	}

	return nil
}
