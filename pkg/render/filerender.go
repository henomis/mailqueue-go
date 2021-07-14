package render

import (
	"io"
	"io/ioutil"
	"os"
)

//FileRender implementation with file
type FileRender struct {
	Path string
}

//Set implemetation
func (fr *FileRender) Set(k Key, v Value) error {

	f, err := os.Create(fr.Path + string(k))
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write([]byte(v))
	if err != nil {
		return err
	}

	return nil
}

//Get implemetation
func (fr *FileRender) Get(k Key) (Value, error) {
	f, err := os.Open(fr.Path + string(k))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	return (Value(b)), nil

}

//Execute implemetation
func (fr *FileRender) Execute(r io.Reader, w io.Writer, k Key) error {

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

	j, err := createJSON(data)
	if err != nil {
		return err
	}

	err = merge(v, w, j)
	if err != nil {
		return err
	}

	return nil
}
