package proxy

import (
	"io"
	"log"
)

const (
	PanicLevel = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

type Logger struct {
	out      io.Writer
	logLevel int
	loggers  map[int]*log.Logger
}

func NewLogger(out io.Writer) *Logger {
	return &Logger{
		out:     out,
		loggers: make(map[int]*log.Logger),
	}
}

func (l *Logger) initLogger(logLevel int) *log.Logger {
	var logPrefix string
	switch logLevel {
	case PanicLevel:
		logPrefix = "PANIC: "
	case FatalLevel:
		logPrefix = "FATAL: "
	case ErrorLevel:
		logPrefix = "ERROR: "
	case WarnLevel:
		logPrefix = "WARN: "
	case InfoLevel:
		logPrefix = "INFO: "
	case DebugLevel:
		logPrefix = "DEBUG: "
	case TraceLevel:
		logPrefix = "TRACE: "
	}
	return log.New(l.out, logPrefix, log.Ldate|log.Ltime|log.Lshortfile)
}

func (l *Logger) log(logLevel int, format string, args ...interface{}) {
	if l.logLevel < logLevel {
		return
	}
	logger := l.loggers[logLevel]
	if logger == nil {
		logger = l.initLogger(logLevel)
	}
	if logger != nil {
		if format == "" {
			logger.Println(args)
		} else {
			logger.Printf(format, args)
		}
	}
}

func (l *Logger) Panic(args ...interface{}) { l.log(PanicLevel, "", args) }
func (l *Logger) Fatal(args ...interface{}) { l.log(FatalLevel, "", args) }
func (l *Logger) Error(args ...interface{}) { l.log(ErrorLevel, "", args) }
func (l *Logger) Warn(args ...interface{})  { l.log(WarnLevel, "", args) }
func (l *Logger) Info(args ...interface{})  { l.log(InfoLevel, "", args) }
func (l *Logger) Debug(args ...interface{}) { l.log(DebugLevel, "", args) }
func (l *Logger) Trace(args ...interface{}) { l.log(TraceLevel, "", args) }

func (l *Logger) Panicf(format string, args ...interface{}) { l.log(PanicLevel, format, args) }
func (l *Logger) Fatalf(format string, args ...interface{}) { l.log(FatalLevel, format, args) }
func (l *Logger) Errorf(format string, args ...interface{}) { l.log(ErrorLevel, format, args) }
func (l *Logger) Warnf(format string, args ...interface{})  { l.log(WarnLevel, format, args) }
func (l *Logger) Infof(format string, args ...interface{})  { l.log(InfoLevel, format, args) }
func (l *Logger) Debugf(format string, args ...interface{}) { l.log(DebugLevel, format, args) }
func (l *Logger) Tracef(format string, args ...interface{}) { l.log(TraceLevel, format, args) }
