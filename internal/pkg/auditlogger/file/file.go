package fileauditlogger

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/henomis/mailqueue-go/internal/pkg/auditlogger"
)

type FileAuditLogger struct {
	logger   *log.Logger
	logLevel auditlogger.Mode
}

func New(outputWriter io.Writer, logLevel auditlogger.Mode) *FileAuditLogger {

	fileAuditLogger := &FileAuditLogger{}
	fileAuditLogger.logger = &log.Logger{}
	fileAuditLogger.logger.SetOutput(outputWriter)
	fileAuditLogger.logLevel = logLevel

	return fileAuditLogger

}

func (f *FileAuditLogger) Log(mode auditlogger.Mode, format string, v ...interface{}) {

	if f == nil {
		return
	}

	if mode < f.logLevel {
		fmt.Printf("livello \n")
		return
	}

	modeStr := fmt.Sprintf("[%s]", mode.String())
	date := time.Now().UTC().Format("2006-01-02 15:04:05")
	str := mode.Color()(fmt.Sprintf("%s %s %s", modeStr, date, format))

	f.logger.Printf(str+"\n", v...)
}
