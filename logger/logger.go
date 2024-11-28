package logger

import (
	"fmt"
	"gosol/monitor"
	"runtime"
	"strings"
)

const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	NoLevel
	TraceLevel
)

var (
	colorOff    = []byte("\033[0m")
	colorRed    = []byte("\033[0;31m")
	colorGreen  = []byte("\033[0;32m")
	colorOrange = []byte("\033[0;33m")
	colorPurple = []byte("\033[0;35m")
	colorCyan   = []byte("\033[0;36m")
)

type Logger struct {
	Level   int
	Prefix  string
	nocolor bool
	logChan chan monitor.StatusMessage
}

func (l *Logger) NoColor() {
	l.nocolor = true
}

func (l *Logger) Color() {
	l.nocolor = false
}

func (l *Logger) colorize(color []byte, s string) string {
	if l.nocolor {
		return s
	}
	return string(color) + s + string(colorOff)
}

func (l *Logger) Lev() string {
	switch l.Level {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case NoLevel:
		return "disabled"
	case TraceLevel:
		return "trace"
	default:
		return "info"
	}
}

func (l *Logger) SetLevel(level string) *Logger {
	switch level {
	case "debug":
		l.Level = DebugLevel
	case "info":
		l.Level = InfoLevel
	case "warn":
		l.Level = WarnLevel
	case "error":
		l.Level = ErrorLevel
	case "disabled":
		l.Level = NoLevel
	default:
		l.Level = InfoLevel
	}
	return l
}

func (l *Logger) logMessage(level string, color []byte, v ...any) {
	if l.logChan != nil {
		message := fmt.Sprintf("%s %s - %s", l.colorize(color, level), l.Prefix, getVariable(v...))
		var statusLevel monitor.LogLevel
		switch level {
		case "[error]":
			statusLevel = monitor.ERR
		case "[warn]":
			statusLevel = monitor.WARN
		case "[info]":
			statusLevel = monitor.INFO
		case "[debug]":
			statusLevel = monitor.DEBUG
		case "[trace]":
			statusLevel = monitor.TRACE
		default:
			statusLevel = monitor.NONE
		}
		l.logChan <- monitor.StatusMessage{Level: statusLevel, Message: message}
	}
}

func (l *Logger) Error(v ...any) {
	if l.Level <= ErrorLevel {
		l.logMessage("[error]", colorRed, v...)
	}
}

func (l *Logger) Warn(v ...any) {
	if l.Level <= WarnLevel {
		l.logMessage("[warn]", colorOrange, v...)
	}
}

func (l *Logger) Info(v ...any) {
	if l.Level <= InfoLevel {
		l.logMessage("[info]", colorGreen, v...)
	}
}

func (l *Logger) Debug(v ...any) {
	if l.Level <= DebugLevel {
		l.logMessage("[debug]", colorPurple, v...)
	}
}

func (l *Logger) Trace(v ...any) {
	if l.Level <= TraceLevel {
		l.logMessage("[trace]", colorCyan, v...)
	}
}

func (l *Logger) Panic(v ...any) {
	stack := make([]byte, 1536)
	runtime.Stack(stack, false)
	if l.logChan != nil {
		message := fmt.Sprintf("%s %s - %s\n%s", l.colorize(colorCyan, "[panic]"), l.Prefix, getVariable(v...), l.colorize(colorOrange, string(stack)))
		l.logChan <- monitor.StatusMessage{Level: monitor.PANIC, Message: message}
	}
}

func NewLogger(prefix string, logChan chan monitor.StatusMessage) *Logger {
	return &Logger{
		Prefix:  prefix,
		logChan: logChan,
	}
}

func getVariable(v ...any) string {
	if len(v) == 0 {
		return ""
	}
	if len(v) == 1 {
		return fmt.Sprint(v[0])
	}
	return strings.Trim(fmt.Sprint(v...), "[]")
}
