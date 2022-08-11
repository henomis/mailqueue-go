package mongostorage

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func newMongoClient(endpoint string) (*mongo.Client, error) {
	mongoClientOptions := options.Client().ApplyURI(endpoint)
	err := mongoClientOptions.Validate()
	if err != nil {
		return nil, errors.Wrap(err, "invalid mongodb endpoint")
	}

	return mongo.NewClient(mongoClientOptions)
}

func mongoCappedSizeFromUint64(cappedSize uint64) *int64 {
	cappedSizeAsInt64 := int64(cappedSize)
	return &cappedSizeAsInt64
}

func (ms *MongoStorage) setupTailableAwaitCursor(filterQuery MongoQuery) error {

	if ms.mongoCursor != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), ms.timeout)
	defer cancel()

	mongoFindOptions := options.Find().SetCursorType(options.TailableAwait).SetNoCursorTimeout(true)

	mongoCursor, err := ms.mongoCollection.Find(ctx, filterQuery, mongoFindOptions)
	if err != nil {
		return err
	}

	ms.mongoCursor = mongoCursor

	return nil
}

func (ms *MongoStorage) waitCursor(filterQuery MongoQuery) error {

	for {

		isNextDocumentAvailable := ms.mongoCursor.TryNext(context.Background())
		if isNextDocumentAvailable {
			break
		} else if ms.mongoCursor.ID() == 0 {
			//empty collection
			time.Sleep(1 * time.Second)
			ms.mongoCursor = nil
			ms.setupTailableAwaitCursor(filterQuery)
			continue
		} else if err := ms.mongoCursor.Err(); err != nil {
			ms.mongoCursor = nil
			return err
		} else {
			// waiting element
			time.Sleep(1 * time.Second)
		}
	}

	return nil
}
