package main

import (
	"bytes"
	"io"
	"log"
	"strings"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/render/mongorender"
)

const (
	key = "template1"
)

func main() {

	render, err := mongorender.New(
		&mongorender.MongoTemplateOptions{
			Endpoint:   "mongodb://localhost:27017",
			Database:   "test",
			Collection: "template",
			Timeout:    time.Second * 5,
		},
	)

	// render, err := filerender.New("./templates")

	if err != nil {
		panic(err)
	}

	err = render.Set(key, "<html><body>Hello {{.nome}}</body></html>")
	if err != nil {
		panic(err)
	}

	value, err := render.Get(key)
	if err != nil {
		panic(err)
	}
	log.Println("Key ", key, " Value: ", value)

	templateData := `{"nome": "Mr. Winston"}`
	var output bytes.Buffer
	bufferWriter := io.Writer(&output)

	render.Execute(strings.NewReader(templateData), bufferWriter, "template1")

	log.Println("Render: ", output.String())

}
