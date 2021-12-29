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

	fileAuditLogger := &FileAuditLogger{}
	fileAuditLogger.logger = &log.Logger{}
	fileAuditLogger.logger.SetOutput(*outputWriter)

	return fileAuditLogger

}

//Trace implementation
func (f *FileAuditLogger) Trace(mode auditlogger.Mode, format string, v ...interface{}) {

	if f == nil {
		return
	}
	f.logger.Printf("["+strings.ToUpper(string(mode))+"] "+format+"\n", v...)
}
