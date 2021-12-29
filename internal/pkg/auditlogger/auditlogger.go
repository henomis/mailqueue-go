package auditlogger

//Mode enum type
type Mode string

const (
	//Info log
	Info = "INFO"
	//Warning log
	Warning = "WARNING"
	//Error log
	Error = "ERROR"
)

//Tracer interface
type AuditLogger interface {
	Trace(Mode, string, ...interface{})
}
