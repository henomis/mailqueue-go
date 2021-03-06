package render

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/henomis/mailqueue-go/pkg/render"
)

//FileRender implementation with file
type FileRender struct {
	Path string
}

//Set implemetation
func (fr *FileRender) Set(k string, v interface{}) error {

	f, err := os.Create(fr.Path + string(k))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(v.(string)))
	if err != nil {
		return err
	}

	return nil
}

//Get implemetation
func (fr *FileRender) Get(k string) (interface{}, error) {
	f, err := os.Open(fr.Path + string(k))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return (string(b)), nil

}

//Execute implemetation
func (fr *FileRender) Execute(r io.Reader, w io.Writer, k string) error {

	//Get template
	v, err := fr.Get(k)
	if err != nil {
		return err
	}

	//Get JSON
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	j, err := render.CreateTemplateDataObject(data)
	if err != nil {
		return err
	}

	err = render.Merge(v.(string), j, w)
	if err != nil {
		return err
	}

	return nil
}
