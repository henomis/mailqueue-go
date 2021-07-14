package render

import (
	"bytes"
	"reflect"
	"testing"
	"time"
)

func TestRender(t *testing.T) {

	k := Key("test")
	v := Value(`<html><body>Hello, {{.world}}</body></html>`)

	checkDeep := func(t testing.TB, got, want []byte) {
		t.Helper()

		if !reflect.DeepEqual(got, want) {
			t.Errorf("Expected %q got %q", want, got)
		}

	}

	tests := []struct {
		name string
		rend Render
	}{
		{
			"file",
			&FileRender{
				Path: "/tmp/",
			},
		},
		{
			"mongo",
			&MongoRender{
				Database: "test",
				Endpoint: "mongodb://admin:pass@localhost:27017/admin?authSource=admin", //os.Getenv("MONGO_ENDPOINT"), //
				Timeout:  10 * time.Second,
			},
		},
	}

	for _, test := range tests {

		t.Run("test Set"+test.name, func(t *testing.T) {

			t.Helper()

			err := test.rend.Set(k, v)
			if err != nil {
				t.Errorf(err.Error())
			}

		})

		t.Run("test Get"+test.name, func(t *testing.T) {

			value, err := test.rend.Get(k)
			if err != nil {
				t.Errorf(err.Error())
			}

			/*if !reflect.DeepEqual(v, value) {
				t.Errorf("Expected %q got %q", v, value)
			}*/

			checkDeep(t, value, v)

		})

		t.Run("test Execute"+test.name, func(t *testing.T) {

			body := []byte(`{"world":"world!"}`)
			want := []byte("<html><body>Hello, world!</body></html>")
			got := []byte{}

			bufferBody := bytes.NewBuffer(body)
			buffGot := bytes.NewBuffer(got)

			err := test.rend.Execute(bufferBody, buffGot, k)
			if err != nil {
				t.Errorf(err.Error())
			}

			got = buffGot.Bytes()

			/*if !reflect.DeepEqual(want, got) {
				t.Errorf("Expected %q got %q", want, got)
			}*/

			checkDeep(t, got, want)

		})

	}

}
