package mongorender

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func createMongoClient(mongoEndpoint string) (*mongo.Client, error) {
	mongoClientOptions := options.Client().ApplyURI(mongoEndpoint)
	err := mongoClientOptions.Validate()
	if err != nil {
		return nil, err
	}

	mongoClient, err := mongo.NewClient(mongoClientOptions)
	if err != nil {
		return nil, err
	}
	return mongoClient, nil
}

func setupMongoConnection(mongoClient *mongo.Client) error {
	err := mongoClient.Connect(context.Background())
	if err != nil {
		return err
	}

	err = mongoClient.Ping(context.Background(), readpref.Primary())
	if err != nil {
		return err
	}

	return nil
}

func createMongoConnection(mongoRender *MongoRender) error {

	mongoClient, err := createMongoClient(mongoRender.MongoEndpoint)
	if err != nil {
		return err
	}
	mongoRender.mongoClient = mongoClient

	err = setupMongoConnection(mongoRender.mongoClient)
	if err != nil {
		return err
	}

	return nil
}

func (mr *MongoRender) selectDatabaseAndCollection() {

	mongoDatabase := mr.mongoClient.Database(mr.MongoDatabase)
	mr.mongoCollection = mongoDatabase.Collection(collectionTemplate)

}
