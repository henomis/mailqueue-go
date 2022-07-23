package main

import (
	"fmt"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/email"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemaillog"
)

// type Item struct {
// 	UUID      string `bson:"_id"`
// 	Service   string `json:"appname" bson:"service"`
// 	Timestmap int64  `json:"timestamp" bson:"timestamp"`
// 	EmailID   string `json:"uuid" bson:"email_id"`
// 	Status    int    `json:"status" bson:"status"`
// 	Error     string `json:"error" bson:"error"`
// }

func main() {
	log, err := mongoemaillog.New(&mongoemaillog.MongoEmailLogOptions{
		Endpoint:   "mongodb://localhost:27017",
		Database:   "test",
		Collection: "log",
		Timeout:    time.Second * 5,
	})
	if err != nil {
		panic(err)
	}

	for i := 0; i < 10; i++ {
		id, err := log.Log(
			&email.Log{
				ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
				Service:   "test",
				Timestamp: time.Now().UTC(),
				EmailID:   "test",
				Status:    1,
				Error:     "test",
			},
		)
		fmt.Println(id, err)
		time.Sleep(time.Millisecond)
	}

	items, err := log.Items("test")
	fmt.Println(items, err)
	// log.Model(&Item{})

	// for i := 0; i < 10; i++ {
	// 	_, err = log.Log(&Item{
	// 		UUID:      fmt.Sprintf("uuid-%d", i),
	// 		Service:   "test",
	// 		Timestmap: time.Now().Unix(),
	// 		EmailID:   fmt.Sprintf("email-%d", 10),
	// 		Status:    i,
	// 		Error:     "",
	// 	})
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	// items, err := log.Items("email-10")
	// if err != nil {
	// 	panic(err)
	// }

	// for _, item := range *items.(*[]Item) {
	// 	fmt.Printf("%#v\n", ItemModel(item))
	// }

}
