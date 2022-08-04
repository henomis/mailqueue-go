package mongotemplate

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/mongostorage"
	"github.com/henomis/mailqueue-go/internal/pkg/render"
	"github.com/henomis/mailqueue-go/internal/pkg/storagemodel"
	"github.com/pkg/errors"
)

type MongoTemplateOptions struct {
	Endpoint   string
	Database   string
	Collection string
	Timeout    time.Duration
}

type MongoTemplate struct {
	mongoTemplateOptions *MongoTemplateOptions
	mongoStorage         *mongostorage.MongoStorage
}

func New(mongoTemplateOptions *MongoTemplateOptions) (*MongoTemplate, error) {
	err := validateMongoTemplateOptions(mongoTemplateOptions)
	if err != nil {
		return nil, errors.Wrap(err, "invalid mongo email log options")
	}

	mongoStorage, err := mongostorage.New(
		mongoTemplateOptions.Endpoint,
		mongoTemplateOptions.Timeout,
		mongoTemplateOptions.Database,
		mongoTemplateOptions.Collection,
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

	return &MongoTemplate{
		mongoTemplateOptions: mongoTemplateOptions,
		mongoStorage:         mongoStorage,
	}, nil

}

// func (mr *MongoTemplate) Set(key string, value interface{}) error {

// 	filterQuery := mongostorage.Queryf(`{"_id": "%s"}`, key)
// 	mongoTemplate := storagemodel.Template{
// 		TemplateIDAndName: storagemodel.TemplateIDAndName{
// 			ID:   key,
// 			Name: key,
// 			// Template: value.(string),
// 		},
// 	}
// 	err := mr.mongoStorage.ReplaceOrInsert(filterQuery, mongoTemplate)
// 	if err != nil {
// 		return errors.Wrap(err, "unable to replace or insert value")
// 	}

// 	return nil
// }

func (mr *MongoTemplate) Get(key string) (interface{}, error) {

	var mongoTemplate storagemodel.Template

	filterQuery := mongostorage.Queryf(`{"_id": "%s"}`, key)
	err := mr.mongoStorage.FindOne(filterQuery, &mongoTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "unable find template")
	}

	return mongoTemplate.Template, nil
}

//Execute implemetation
func (mr *MongoTemplate) Execute(inputDataReader io.Reader, outputDataWriter io.Writer, mongoKey string) error {

	mongoTemplateBody, err := mr.Get(mongoKey)
	if err != nil {
		return errors.Wrapf(err, "unable to get mongo template %s", mongoKey)
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

// CRUD
func (mr *MongoTemplate) Create(mongoTemplate *storagemodel.Template) (string, error) {

	mongoTemplate.ID = mongostorage.RandomID()

	templateID, err := mr.mongoStorage.InsertOne(mongoTemplate)
	if err != nil {
		return "", errors.Wrap(err, "unable insert template")
	}

	return templateID.(string), nil
}

func (mr *MongoTemplate) Read(id string) (*storagemodel.Template, error) {
	var mongoTemplate storagemodel.Template

	filterQuery := mongostorage.Queryf(`{"_id": "%s"}`, id)
	err := mr.mongoStorage.FindOne(filterQuery, &mongoTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "unable find template")
	}

	return &mongoTemplate, nil
}

func (mr *MongoTemplate) ReadAll(limit, skip int64, fields string) ([]storagemodel.Template, int64, error) {
	var mongoTemplates []storagemodel.Template

	findOptions := mongostorage.SetLimit(nil, limit)
	findOptions = mongostorage.SetSkip(findOptions, skip)
	if len(fields) > 0 {
		fieldsParts := strings.Split(fields, ",")
		findOptions = mongostorage.SetProjection(nil, fieldsParts)
	}

	count, err := mr.mongoStorage.CountQuery(mongostorage.Query(""))
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable count templates")
	}

	err = mr.mongoStorage.DecodeAll(mongostorage.Query(""), findOptions, &mongoTemplates)
	if err != nil {
		return nil, 0, errors.Wrap(err, "unable find templates")
	}

	return mongoTemplates, count, nil
}

func (mr *MongoTemplate) Update(id string, mongoTemplate *storagemodel.Template) error {
	filterQuery := mongostorage.Queryf(`{"_id": "%s"}`, id)
	updateQuery := mongostorage.Queryf(`{"$set": {"template": "%s", "name":"%s"}}`,
		mongoTemplate.Template,
		mongoTemplate.Name,
	)

	err := mr.mongoStorage.Update(filterQuery, updateQuery)
	if err != nil {
		return errors.Wrap(err, "unable to update data")
	}

	return err
}

func (mr *MongoTemplate) Delete(id string) error {
	filterQuery := mongostorage.Queryf(`{"_id": "%s"}`, id)
	err := mr.mongoStorage.DeleteOne(filterQuery)
	if err != nil {
		return errors.Wrap(err, "unable delete template")
	}

	return nil
}

// ---------------
// Support methods
// ---------------

func validateMongoTemplateOptions(mongoTemplateOptions *MongoTemplateOptions) error {

	if len(mongoTemplateOptions.Endpoint) == 0 {
		return fmt.Errorf("invalid endpoint")
	}

	if len(mongoTemplateOptions.Database) == 0 {
		return fmt.Errorf("invalid database name")
	}

	if len(mongoTemplateOptions.Collection) == 0 {
		return fmt.Errorf("invalid collection name")
	}

	return nil
}
