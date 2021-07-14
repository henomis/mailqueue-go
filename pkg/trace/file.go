package trace

import (
	"log"
	"os"
	"strings"
)

const (
	filePrefix = "file://"
)

//FileTracer struct
type FileTracer struct {
	logger *log.Logger
}

type nullWriter struct {
}

func (w *nullWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

//NewFileTracer func
func NewFileTracer(output string) *FileTracer {

	t := &FileTracer{}
	t.logger = &log.Logger{}

	if strings.HasPrefix(output, filePrefix) {

		file, err := os.OpenFile(strings.TrimPrefix(output, filePrefix), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			panic(err)
		}
		t.logger.SetOutput(file)
	} else if output == "-" {
		t.logger.SetOutput(&nullWriter{})
	} else {
		t.logger.SetOutput(os.Stdout)
	}

	return t

}

//Trace implementation
func (t *FileTracer) Trace(mode Mode, format string, v ...interface{}) {

	if t == nil {
		return
	}
	t.logger.Printf("["+strings.ToUpper(string(mode))+"] "+format+"\n", v...)
}
