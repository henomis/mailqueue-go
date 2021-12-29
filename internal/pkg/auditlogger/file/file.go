package fileauditlogger

import (
	"io"
	"log"
	"strings"

	"github.com/henomis/mailqueue-go/internal/pkg/auditlogger"
)

const (
	filePrefix = "file://"
)

//FileAuditLogger struct
type FileAuditLogger struct {
	logger *log.Logger
}

//NewFileAuditLogger func
func NewFileAuditLogger(outputWriter *io.Writer) *FileAuditLogger {

	t := &FileAuditLogger{}
	t.logger = &log.Logger{}
	t.logger.SetOutput(*outputWriter)

	return t

}

//Trace implementation
func (t *FileAuditLogger) Trace(mode auditlogger.Mode, format string, v ...interface{}) {

	if t == nil {
		return
	}
	t.logger.Printf("["+strings.ToUpper(string(mode))+"] "+format+"\n", v...)
}
