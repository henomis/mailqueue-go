package render

import (
	"encoding/json"
	"io"
	"reflect"
	"text/template"
)

//Render default implementation
func Merge(templateBody string, templateDataObject map[string]interface{}, outputDataWriter io.Writer) error {

	templateFuncMap := template.FuncMap{
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
	tmplRender := template.New("mailTemplate").Funcs(templateFuncMap)

	parsedTemplate, err := tmplRender.Parse(templateBody)
	if err != nil {
		return err
	}

	if err = parsedTemplate.Execute(outputDataWriter, &templateDataObject); err != nil {
		return err
	}

	return nil
}

//CreateTemplateDataObject utils
func CreateTemplateDataObject(templateData []byte) (map[string]interface{}, error) {

	funcMap := make(map[string]interface{})
	if err := json.Unmarshal(templateData, &funcMap); err != nil {
		return nil, err
	}

	return funcMap, nil

}
