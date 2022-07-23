package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/email"
	"github.com/henomis/mailqueue-go/internal/pkg/limiter"
	"github.com/henomis/mailqueue-go/internal/pkg/mongoemailqueue"
)

// type Item struct {
// 	UUID string `bson:"_id"`
// 	Name string `bson:"name"`
// 	Sent bool   `bson:"sent"`
// }

// func jsonTag(v interface{}, fieldName string) (string, bool) {

// 	fieldNameParts := strings.Split(fieldName, ".")

// 	if len(fieldNameParts) > 1 {
// 		field := reflect.ValueOf(v).FieldByName(fieldNameParts[0])
// 		if field.Kind() == reflect.Struct {
// 			return jsonTag(field.Interface(), fieldNameParts[1])
// 		}
// 	} else {
// 		sf, ok := reflect.TypeOf(v).FieldByName(fieldName)
// 		if !ok {
// 			return "", false
// 		}
// 		return sf.Tag.Lookup("json")
// 	}

// 	return "", false

// }

func main() {

	q, err := mongoemailqueue.New(
		&mongoemailqueue.MongoEmailQueueOptions{
			Endpoint:   "mongodb://localhost:27017",
			Database:   "test",
			Collection: "test",
			CappedSize: 1000000,
			Timeout:    time.Second * 5,
		},
		limiter.NewFixedWindowLimiter(
			3,
			time.Second*2,
		),
		nil,
	)
	if err != nil {
		panic(err)
	}

	// 	q.Model(&Item{})

	go func(q *mongoemailqueue.MongoEmailQueue) {
		// time.Sleep(time.Second * 5)
		for {

			item, err := q.Dequeue()
			if err != nil {
				log.Println(err)
				time.Sleep(time.Second * 1)
				continue
			}

			j, _ := json.MarshalIndent(item, "", "  ")
			log.Println(string(j))

			//q.Update(ItemModel(item).UUID, "sent", true)
			q.SetProcessed(item.ID)

		}
	}(q)

	// lastItemID := ""
	// for i := 0; i < 3; i++ {
	// 	item := &Item{
	// 		UUID: fmt.Sprintf("%d", time.Now().Unix()),
	// 		Name: fmt.Sprintf("Jhon %d", time.Now().Unix()),
	// 		Sent: false,
	// 	}
	// 	q.Enqueue(item)
	// 	lastItemID = item.UUID

	// 	time.Sleep(1 * time.Second)
	// }

	// 	time.Sleep(3 * time.Second)

	// 	item := &Item{
	// 		UUID: fmt.Sprintf("%d", time.Now().Unix()),
	// 		Name: fmt.Sprintf("Jhon %d", time.Now().Unix()),
	// 		Sent: false,
	// 	}
	// 	q.Enqueue(item)

	for i := 0; i < 10; i++ {
		fmt.Println(q.Enqueue(&email.Email{
			ID:          fmt.Sprintf("%d-%d", time.Now().UnixNano(), i),
			Service:     "service",
			To:          "no-reply@example.com",
			Subject:     "subject",
			Cc:          "cc",
			Bcc:         "bcc",
			HTML:        "<h1>Hello</h1>",
			Data:        "data",
			Attachments: []email.Attachment{},
			Template:    "template",
			Processed:   false,
			Status:      email.StatusQueued,
		}))
		time.Sleep(1 * time.Millisecond)

	}

	time.Sleep(30 * time.Second)

	// 	// selectOne, _ := q.Query(fmt.Sprintf(`{"_id":"%s"}`, lastItemID))
	// 	// var item1 Item
	// 	// q.Select(selectOne, &item1)

	// 	// fmt.Printf("%#v\n", item1)
}

// func ItemModel(item interface{}) *Item {
// 	return item.(*Item)
// }
