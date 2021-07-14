package render

import (
	"encoding/json"
	"html/template"
	"io"
	"reflect"
)

//Key render
type Key string

//Value render
type Value []byte

type jsonData map[string]interface{}

//Render interface
type Render interface {
	Set(Key, Value) error
	Get(Key) (Value, error)
	Execute(io.Reader, io.Writer, Key) error
}

//Render default implementation
func merge(data []byte, w io.Writer, json jsonData) error {

	tf := template.FuncMap{
		"isInt": func(i interface{}) bool {
			v := reflect.ValueOf(i)
			switch v.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64, reflect.Float32, reflect.Float64:
				return true
			default:
				return false
			}
		},
		"isString": func(i interface{}) bool {
			v := reflect.ValueOf(i)
			switch v.Kind() {
			case reflect.String:
				return true
			default:
				return false
			}
		},
		"isSlice": func(i interface{}) bool {
			v := reflect.ValueOf(i)
			switch v.Kind() {
			case reflect.Slice:
				return true
			default:
				return false
			}
		},
		"isArray": func(i interface{}) bool {
			v := reflect.ValueOf(i)
			switch v.Kind() {
			case reflect.Array:
				return true
			default:
				return false
			}
		},
		"isMap": func(i interface{}) bool {
			v := reflect.ValueOf(i)
			switch v.Kind() {
			case reflect.Map:
				return true
			default:
				return false
			}
		},
	}
	tmplRender := template.New("mailTemplate").Funcs(tf)

	tt, err := tmplRender.Parse(string(data))
	if err != nil {
		return err
	}

	if err = tt.Execute(w, &json); err != nil {
		return err
	}

	return nil
}

//CreateJSON utils
func createJSON(template Value) (jsonData, error) {

	m := make(map[string]interface{})
	if err := json.Unmarshal(template, &m); err != nil {
		return nil, err
	}

	return m, nil

}
