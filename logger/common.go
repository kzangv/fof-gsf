package logger

const (
	Debug = iota
	Info
	Warn
	Error
)

type Interface interface {
	Level() int
	SetLevel(int) int

	Debug(format string, v ...interface{})
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})

	DebugForce(format string, v ...interface{})
	InfoForce(format string, v ...interface{})
	WarnForce(format string, v ...interface{})
	ErrorForce(format string, v ...interface{})
}
