package auditlogger

type Mode string

const (
	//Info log
	Info = "INFO"
	//Warning log
	Warning = "WARNING"
	//Error log
	Error = "ERROR"
)

type AuditLogger interface {
	Log(Mode, string, ...interface{})
}
