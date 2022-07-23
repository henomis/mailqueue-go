package audit

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

var logger *log.Logger
var level Mode

func init() {
	logger = &log.Logger{}
	logger.SetOutput(os.Stdout)
	level = Info
}

func SetLevel(mode Mode) {
	level = mode
}

func SetOutput(outputWriter io.Writer) {
	logger.SetOutput(outputWriter)
}

func Log(mode Mode, format string, v ...interface{}) {

	if mode < level {
		fmt.Printf("livello \n")
		return
	}

	modeStr := fmt.Sprintf("[%s]", mode.String())
	date := time.Now().UTC().Format("2006-01-02 15:04:05")
	str := mode.Color()(fmt.Sprintf("%s %s %s", modeStr, date, format))

	logger.Printf(str+"\n", v...)
}
