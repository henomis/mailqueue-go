package mongowatchablequeue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Support methods

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

func createMongoConnection(mongoQueue *MongoWatchableQueue, mongoOptions *MongoWatchableQueueOptions) error {

	mongoClient, err := createMongoClient(mongoOptions.MongoEndpoint)
	if err != nil {
		return err
	}
	mongoQueue.mongoClient = mongoClient

	err = setupMongoConnection(mongoQueue.mongoClient)
	if err != nil {
		return err
	}

	return nil
}

func (q *MongoWatchableQueue) selectDatabaseAndCollection(mongoOptions *MongoWatchableQueueOptions) error {
	q.mongoDatabase = q.mongoClient.Database(mongoOptions.MongoDatabase)

	err := q.createCappedCollectionIfNotExists(mongoOptions.MongoCollection, &mongoOptions.MongoCappedSize)
	if err != nil {
		return fmt.Errorf("createCappedCollectionIfNotExists(mongoCollection: %s, mongoCappedSize: %d): %w", mongoOptions.MongoCollection, mongoOptions.MongoCappedSize, err)
	}

	q.mongoCollection = q.mongoDatabase.Collection(mongoOptions.MongoCollection)

	return nil

}

func (q *MongoWatchableQueue) setupMongoFilterAndUpdateCommit(mongoOptions *MongoWatchableQueueOptions) error {

	err := json.Unmarshal([]byte(mongoOptions.MongoDocumentFilter), &q.MongoDocumentFilter)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(mongoOptions.MongoUpdateOnCommit), &q.MongoUpdateOnCommit)
	if err != nil {
		return err
	}

	return nil
}

func (q *MongoWatchableQueue) createCappedCollectionIfNotExists(mongoCappedCollection string, mongoCappedCollectionSize *int64) error {

	mongoCollections, err := q.mongoDatabase.ListCollectionNames(context.Background(), bson.D{}, nil)
	if err != nil {
		return err
	}

	for _, mongoCollection := range mongoCollections {
		if mongoCollection == mongoCappedCollection {
			return nil
		}
	}

	isCapped := true
	cc := &options.CreateCollectionOptions{
		Capped:      &isCapped,
		SizeInBytes: mongoCappedCollectionSize,
	}

	err = q.mongoDatabase.CreateCollection(context.Background(), mongoCappedCollection, cc)
	if err != nil {
		return err
	}

	return nil
}

func (q *MongoWatchableQueue) setTailableMongoCursor() error {

	var err error

	mongoCollectionCursorOptions := options.Find().SetCursorType(options.TailableAwait)
	q.mongoCollectionCursor, err = q.mongoCollection.Find(context.Background(), q.MongoDocumentFilter, mongoCollectionCursorOptions)

	return err
}

func (q *MongoWatchableQueue) waitNextMongoDocument() error {

	if q.mongoCollectionCursor == nil {

		err := q.setTailableMongoCursor()
		if err != nil {
			q.mongoCollectionCursor = nil
			return err
		}
	}

	//check empty collection
	for {

		isNextMongoDocumentAvailable := q.mongoCollectionCursor.TryNext(context.Background())
		if isNextMongoDocumentAvailable {
			break
		} else if q.mongoCollectionCursor.ID() == 0 {
			//empty collection
			time.Sleep(1 * time.Second)

			q.setTailableMongoCursor()
			continue

		} else if err := q.mongoCollectionCursor.Err(); err != nil {
			time.Sleep(1 * time.Second)

			q.setTailableMongoCursor()
			continue
		} else {
			// waiting element
			time.Sleep(1 * time.Second)
		}

	}

	return nil

}

func (q *MongoWatchableQueue) closeMongoCollectionCursorChannels() {
	close(q.mongoCollectionCursorChannel)
	q.mongoCollectionCursorChannel = nil
}

func randomObjectID() string {
	return primitive.NewObjectID().Hex()
}
