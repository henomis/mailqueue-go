package render

import (
	"io"
)

//Render interface
type Render interface {
	Set(string, interface{}) error
	Get(string) (interface{}, error)
	Execute(io.Reader, io.Writer, string) error
}
