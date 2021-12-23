package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/henomis/mailqueue-go/pkg/limiter"
	mongowatchablequeue "github.com/henomis/mailqueue-go/pkg/watchablequeue/mongo"
)

type MyDocument struct {
	Name  string `bson:"name"`
	Value int64  `bson:"value"`
	Sent  bool   `bson:"sent"`
}

func (p *MyDocument) String() string {
	return fmt.Sprintf("name: %s value: %d", p.Name, p.Value)
}

func main() {

	mongoCappedSize := os.Getenv("MONGO_CAPPED_SIZE")
	mongoCappedSizeInt, err := strconv.ParseInt(mongoCappedSize, 10, 64)
	if err != nil {
		panic(err)
	}

	limiter := limiter.NewDefaultLimiter(3, 1*time.Minute, &limiter.RealSleeper{})

	q, err := mongowatchablequeue.NewMongoQueue(
		&mongowatchablequeue.MongoWatchableQueueOptions{
			MongoEndpoint:       os.Getenv("MONGO_ENDPOINT"),
			MongoDatabase:       os.Getenv("MONGO_DATABASE"),
			MongoCollection:     os.Getenv("MONGO_COLLECTION"),
			MongoCappedSize:     mongoCappedSizeInt,
			MongoDocumentFilter: `{"value.sent":false}`,
			MongoUpdateOnCommit: `{"$set": {"value.sent": true}}`,
		},
		limiter,
	)

	if err != nil {
		panic(err)
	}

	document1 := MyDocument{
		Name:  "Winston",
		Value: time.Now().Unix(),
		Sent:  false,
	}
	container2 := &mongowatchablequeue.MongoElement{}

	container := &mongowatchablequeue.MongoElement{
		Value: document1,
	}

	ch, err := q.Watch(container2)
	if err != nil {
		panic(err)
	}

	go func() {
		for i := range ch {

			g, ok := i.(*mongowatchablequeue.MongoElement)
			if ok {

				log.Println("dec ", g)
				q.Commit(g)
			}

		}
	}()

	time.Sleep(1 * time.Second)
	log.Println("equeue")
	q.Enqueue(container)

	time.Sleep(10 * time.Second)

}
