package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	mongowatchablequeue "github.com/henomis/mailqueue-go/pkg/watchablequeue/mongo"
)

type Pippo struct {
	Name  string `bson:"name"`
	Value int64  `bson:"value"`
	Sent  bool   `bson:"sent"`
}

func (p *Pippo) String() string {
	return fmt.Sprintf("name: %s value: %d", p.Name, p.Value)
}

func main() {

	mongoCappedSize := os.Getenv("MONGO_CAPPED_SIZE")
	mongoCappedSizeInt, err := strconv.ParseInt(mongoCappedSize, 10, 64)
	if err != nil {
		panic(err)
	}

	q, err := mongowatchablequeue.NewMongoQueue(
		&mongowatchablequeue.MongoWatchableQueueOptions{
			MongoEndpoint:       os.Getenv("MONGO_ENDPOINT"),
			MongoDatabase:       os.Getenv("MONGO_DATABASE"),
			MongoCollection:     os.Getenv("MONGO_COLLECTION"),
			MongoCappedSize:     mongoCappedSizeInt,
			MongoDocumentFilter: `{"value.sent":false}`,
			MongoUpdateOnCommit: `{"$set": {"value.sent": true}}`,
		},
	)

	if err != nil {
		panic(err)
	}

	pippo := Pippo{

		Name:  "pippo3",
		Value: time.Now().Unix(),
		Sent:  false,
	}
	pippo2 := &mongowatchablequeue.MongoElement{}

	container := &mongowatchablequeue.MongoElement{
		Value: pippo,
	}

	ch, err := q.Watch(pippo2)
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
			//			q.Commit()

		}
	}()

	// go func(channel <-chan interface{}) {
	// 	for v := range ch {

	// 		e := v.(*Pippo)
	// 		log.Printf("%+v\n", e)
	// 	}

	// }(ch)

	time.Sleep(1 * time.Second)
	log.Println("equeue")
	q.Enqueue(container)

	time.Sleep(10 * time.Second)

}
