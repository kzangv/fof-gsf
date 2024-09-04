package logger

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"runtime"
)

const (
	// Colors
	_LogColorReset       = "\033[0m"
	_LogColorRedBold     = "\033[1;31m"
	_LogColorGreenBold   = "\033[1;32m"
	_LogColorYellowBold  = "\033[1;33m"
	_LogColorBlueBold    = "\033[1;34m"
	_LogColorMagentaBold = "\033[1;35m"
	_LogColorCyanBold    = "\033[1;36m"
	_LogColorWhiteBold   = "\033[1;37m"
)

type Console struct {
	level int

	stdOutput *log.Logger
	errOutput *log.Logger

	debugStr, infoStr, warnStr, errStr string
}

func (output *Console) Level() int {
	return output.level
}
func (output *Console) SetLevel(level int) int {
	old := output.level
	output.level = level
	return old
}
func (output *Console) Init(level int, colorful bool, normal, err io.Writer) {
	output.level = level
	output.stdOutput = log.New(normal, "", log.Ldate|log.Lmicroseconds)
	output.errOutput = log.New(err, "", log.Ldate|log.Lmicroseconds)

	if colorful {
		output.debugStr = _LogColorWhiteBold + "[Debug] " + _LogColorReset
		output.infoStr = _LogColorBlueBold + "[Info] " + _LogColorReset
		output.warnStr = _LogColorMagentaBold + "[Warn] " + _LogColorReset
		output.errStr = _LogColorRedBold + "[Error] " + _LogColorReset
	} else {
		output.infoStr = "[Debug] "
		output.infoStr = "[Info] "
		output.warnStr = "[Warn] "
		output.errStr = "[Error] "
	}
}

func (output *Console) Debug(format string, v ...interface{}) {
	if Debug >= output.level {
		output.DebugForce(format, v...)
	}
}
func (output *Console) Info(format string, v ...interface{}) {
	if Info >= output.level {
		output.InfoForce(format, v...)
	}
}
func (output *Console) Warn(format string, v ...interface{}) {
	if Warn >= output.level {
		output.WarnForce(format, v...)
	}
}
func (output *Console) Error(format string, v ...interface{}) {
	if Error >= output.level {
		output.ErrorForce(format, v...)
	}
}

func (output *Console) DebugForce(format string, v ...interface{}) {
	var buff bytes.Buffer
	buff.WriteString(output.debugStr)
	buff.WriteString(format)
	output.stdOutput.Printf(buff.String(), v...)
}
func (output *Console) InfoForce(format string, v ...interface{}) {
	var buff bytes.Buffer
	buff.WriteString(output.infoStr)
	buff.WriteString(format)
	output.stdOutput.Printf(buff.String(), v...)
}
func (output *Console) WarnForce(format string, v ...interface{}) {
	var buff bytes.Buffer
	buff.WriteString(output.warnStr)
	buff.WriteString(format)
	output.stdOutput.Printf(buff.String(), v...)
}
func (output *Console) ErrorForce(format string, v ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	output.errOutput.Printf(fmt.Sprintf("%s%s\n\tfile: %s:%d", output.errStr, format, file, line), v...)
}
