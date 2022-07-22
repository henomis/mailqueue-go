package mongoemailqueue

// import (
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"log"
// 	"time"

// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// 	"go.mongodb.org/mongo-driver/mongo/readpref"
// )

// func selectID(id interface{}) MongoQueueQuery {
// 	return Query(fmt.Sprintf(`{"_id": "%s"}`, id))
// }

// func (q *MongoQueue) setupMongoCollection() error {
// 	ctx, cancel := context.WithTimeout(context.Background(), q.MongoQueueOptions.Timeout)
// 	defer cancel()

// 	collectionExists := false
// 	db := q.mongoClient.Database(q.MongoQueueOptions.Database)

// 	mongoCollections, err := db.ListCollectionNames(ctx, bson.D{}, nil)
// 	if err != nil {
// 		return fmt.Errorf("unable list collections: %w", err)
// 	}

// 	for _, mongoCollection := range mongoCollections {
// 		if mongoCollection == q.MongoQueueOptions.Collection {
// 			collectionExists = true
// 			break
// 		}
// 	}

// 	if !collectionExists {

// 		isTrue := true
// 		createCollectionOptions := &options.CreateCollectionOptions{
// 			Capped:      &isTrue,
// 			SizeInBytes: mongoCappedSizeFromUint64(q.MongoQueueOptions.CappedSize),
// 		}
// 		err = db.CreateCollection(ctx, q.MongoQueueOptions.Collection, createCollectionOptions)
// 		if err != nil {
// 			return fmt.Errorf("unable create collection: %w", err)
// 		}
// 	}

// 	q.mongoCollection = db.Collection(q.MongoQueueOptions.Collection)

// 	return nil

// }

// func (q *MongoQueue) setupMongoCursor() error {

// 	ctx, cancel := context.WithTimeout(context.Background(), q.MongoQueueOptions.Timeout)
// 	defer cancel()

// 	mongoFindOptions := options.Find().SetCursorType(options.TailableAwait).SetNoCursorTimeout(true)

// 	mongoCursor, err := q.mongoCollection.Find(ctx, q.MongoQueueOptions.Filter, mongoFindOptions)
// 	if err != nil {
// 		return err
// 	}

// 	q.mongoCursor = mongoCursor

// 	return nil
// }

// func (q *MongoQueue) mongoCursorWait() error {

// 	for {

// 		isNextDocumentAvailable := q.mongoCursor.TryNext(context.Background())
// 		if isNextDocumentAvailable {

// 			log.Println("mongo cursor has next document")
// 			break

// 		} else if q.mongoCursor.ID() == 0 {

// 			//empty collection
// 			log.Println("empty collection")
// 			time.Sleep(1 * time.Second)
// 			q.setupMongoCursor()

// 			continue

// 		} else if err := q.mongoCursor.Err(); err != nil {

// 			log.Println("mongo cursor error: ", err)
// 			return q.mongoCursor.Err()

// 		} else {

// 			log.Println("mongo cursor else sleep")
// 			// waiting element
// 			time.Sleep(1 * time.Second)

// 		}
// 	}

// 	return nil
// }

// func mongoCappedSizeFromUint64(cappedSize uint64) *int64 {
// 	cappedSizeAsInt64 := int64(cappedSize)
// 	return &cappedSizeAsInt64
// }
