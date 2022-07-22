package mongorender

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/mongostorage"
	"github.com/henomis/mailqueue-go/internal/pkg/render"
	"github.com/pkg/errors"
)

type MongoTemplate struct {
	ID       string `json:"id" bson:"_id"`
	Template string `json:"template" bson:"template"`
}

type MongoRenderOptions struct {
	Endpoint   string
	Database   string
	Collection string
	Timeout    time.Duration
}

type MongoRender struct {
	mongoRenderOptions *MongoRenderOptions
	mongoStorage       *mongostorage.MongoStorage
}

func New(mongoRenderOptions *MongoRenderOptions) (*MongoRender, error) {
	err := validateMongoRenderOptions(mongoRenderOptions)
	if err != nil {
		return nil, errors.Wrap(err, "invalid mongo email log options")
	}

	mongoStorage, err := mongostorage.New(
		mongoRenderOptions.Endpoint,
		mongoRenderOptions.Timeout,
		mongoRenderOptions.Database,
		mongoRenderOptions.Collection,
		0,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable create mongostorage")
	}

	err = mongoStorage.Connect()
	if err != nil {
		return nil, errors.Wrap(err, "unable connect")
	}

	mongoStorage.CreateCollection()

	return &MongoRender{
		mongoRenderOptions: mongoRenderOptions,
		mongoStorage:       mongoStorage,
	}, nil

}

func (mr *MongoRender) Set(key string, value interface{}) error {

	filterQuery := mongostorage.Queryf(`{_id: "%s"}`, key)
	mongoTemplate := MongoTemplate{
		ID:       key,
		Template: value.(string),
	}
	err := mr.mongoStorage.ReplaceOrInsert(filterQuery, mongoTemplate)
	if err != nil {
		return errors.Wrap(err, "unable to replace or insert value")
	}

	return nil
}

func (mr *MongoRender) Get(key string) (interface{}, error) {

	var mongoTemplate MongoTemplate

	filterQuery := mongostorage.Queryf(`{_id: "%s"}`, key)
	err := mr.mongoStorage.FindOne(filterQuery, &mongoTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "unable find template")
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
		return errors.Wrap(err, "unable read input data")
	}

	templateDataObject, err := render.CreateTemplateDataObject(rawDataFromReader)
	if err != nil {
		return errors.Wrap(err, "unable create template data object")
	}

	err = render.Merge(mongoTemplateBodyAsString, templateDataObject, outputDataWriter)
	if err != nil {
		return errors.Wrap(err, "unable merge template")
	}

	return nil
}

// ---------------
// Support methods
// ---------------

func validateMongoRenderOptions(mongoRenderOptions *MongoRenderOptions) error {

	if len(mongoRenderOptions.Endpoint) == 0 {
		return fmt.Errorf("invalid endpoint")
	}

	if len(mongoRenderOptions.Database) == 0 {
		return fmt.Errorf("invalid database name")
	}

	if len(mongoRenderOptions.Collection) == 0 {
		return fmt.Errorf("invalid collection name")
	}

	return nil
}
