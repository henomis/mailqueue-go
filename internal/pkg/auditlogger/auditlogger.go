package auditlogger

import "fmt"

type Mode int

const (
	Info = iota
	Warning
	Error
)

type AuditLogger interface {
	Log(Mode, string, ...interface{})
}

var modeStrings = map[Mode]string{
	Info:    "INFO",
	Warning: "WARNING",
	Error:   "ERROR",
}

func (m Mode) String() string {
	return modeStrings[m]
}

func (m Mode) Color() func(...interface{}) string {
	return modeColor[m]
}

var modeColor = map[Mode]func(...interface{}) string{
	Info:    color("\033[1;32m%s\033[0m"),
	Warning: color("\033[1;33m%s\033[0m"),
	Error:   color("\033[1;31m%s\033[0m"),
}

func color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}
