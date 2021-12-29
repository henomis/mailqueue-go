package fileauditlogger

import (
	"io"
	"log"
	"strings"

	"github.com/henomis/mailqueue-go/internal/pkg/auditlogger"
)

type FileAuditLogger struct {
	logger *log.Logger
}

func NewFileAuditLogger(outputWriter io.Writer) *FileAuditLogger {

	fileAuditLogger := &FileAuditLogger{}
	fileAuditLogger.logger = &log.Logger{}
	fileAuditLogger.logger.SetOutput(outputWriter)

	return fileAuditLogger

}

func (f *FileAuditLogger) Log(mode auditlogger.Mode, format string, v ...interface{}) {

	if f == nil {
		return
	}
	f.logger.Printf("["+strings.ToUpper(string(mode))+"] "+format+"\n", v...)
}
