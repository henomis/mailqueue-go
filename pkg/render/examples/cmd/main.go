package main

import (
	"bytes"
	"io"
	"log"
	"os"
	"strings"
	"time"

	mongorender "github.com/henomis/mailqueue-go/pkg/render/mongo"
)

const (
	MongoKey = "template1"
)

func main() {

	mongoRender, err := mongorender.NewMongoRender(
		10*time.Second,
		os.Getenv("MONGO_ENDPOINT"),
		os.Getenv("MONGO_DATABASE"),
	)
	if err != nil {
		panic(err)
	}

	err = mongoRender.Set(MongoKey, "<html><body>Hello {{.nome}}</body></html>")
	if err != nil {
		panic(err)
	}

	mongoValue, err := mongoRender.Get(MongoKey)
	if err != nil {
		panic(err)
	}
	log.Println("Key ", MongoKey, " Value: ", mongoValue)

	templateData := `{"nome": "Mr. Winston"}`
	var output bytes.Buffer
	bufferWriter := io.Writer(&output)

	mongoRender.Execute(strings.NewReader(templateData), bufferWriter, "template1")

	log.Println("Render: ", output.String())

}
