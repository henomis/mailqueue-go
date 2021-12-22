package mongowatchablequeue

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoWatchableQueue struct {
	mongoClient                  *mongo.Client
	mongoDatabase                *mongo.Database
	mongoCollection              *mongo.Collection
	mongoCollectionCursor        *mongo.Cursor
	mongoCollectionCursorChannel chan interface{}
	mongoCollectionWatchedFlag   mongoWatchableQueueFlag

	MongoDocumentFilter bson.M
	MongoUpdateOnCommit bson.M
}

func NewMongoQueue(mongoOptions *MongoWatchableQueueOptions) (*MongoWatchableQueue, error) {

	mongoQueue := &MongoWatchableQueue{}

	err := createMongoConnection(mongoQueue, mongoOptions)
	if err != nil {
		return nil, err
	}

	err = mongoQueue.selectDatabaseAndCollection(mongoOptions)
	if err != nil {
		return nil, err
	}

	err = mongoQueue.setupMongoFilterAndUpdateCommit(mongoOptions)
	if err != nil {
		return nil, err
	}

	return mongoQueue, err
}

func (q *MongoWatchableQueue) Enqueue(element interface{}) error {

	mongoElement, err := validateMongoElement(element)
	if err != nil {
		return err
	}
	mongoElement.ID = randomObjectID()

	_, err = q.mongoCollection.InsertOne(context.Background(), mongoElement)
	return err
}

func (q *MongoWatchableQueue) Dequeue(element interface{}) error {

	if q.mongoClient == nil {
		return fmt.Errorf("invalid mongo client")
	}

	err := q.waitNextMongoDocument()
	if err != nil {
		return err
	}

	//waiting limiter
	// for {
	// 	if q.Limiter.Allow() {
	// 		break
	// 	}
	// 	//waiting limiter
	// 	time.Sleep(1 * time.Second)
	// }

	return q.mongoCollectionCursor.Decode(element)

}

func (q *MongoWatchableQueue) Unwatch() {
	q.mongoCollectionWatchedFlag.SetWatched(false)
}

func (q *MongoWatchableQueue) Watch(element interface{}) (<-chan interface{}, error) {

	if q.mongoCollectionCursorChannel != nil {
		return nil, fmt.Errorf("this queue is already watched")
	}

	q.mongoCollectionCursorChannel = make(chan interface{})
	q.mongoCollectionWatchedFlag.SetWatched(true)

	go func(mongoQueue *MongoWatchableQueue, queueElement interface{}) {
		for mongoQueue.mongoCollectionWatchedFlag.IsWatched() {

			err := mongoQueue.Dequeue(queueElement)
			if err != nil {
				mongoQueue.closeMongoCollectionCursorChannels()
				return
			}

			mongoQueue.mongoCollectionCursorChannel <- queueElement

		}
		mongoQueue.closeMongoCollectionCursorChannels()

	}(q, element)

	return q.mongoCollectionCursorChannel, nil

}

//Commit mongodb implementation
func (q *MongoWatchableQueue) Commit(element interface{}) error {

	mongoElement, err := validateMongoElement(element)
	if err != nil {
		return err
	}

	_, err = q.mongoCollection.UpdateOne(context.Background(), bson.M{"_id": mongoElement.ID}, q.MongoUpdateOnCommit)
	if err != nil {
		return err
	}
	return nil
}
