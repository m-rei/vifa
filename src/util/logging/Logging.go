package logging

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"
)

// LogLevel represents a log level...
type LogLevel int

const (
	// All logs pass through
	All LogLevel = iota
	// Debug level logs are just for debugging purposes, where granular information is required
	Debug
	// Info level logs are purely informative
	Info
	// Warn level logs are about unexpected states, which dont hinder core functionality
	Warn
	// Error level represents logs about errors, which break core functionality, but not the whole system
	Error
	// Fatal level errors bring the entire system down
	Fatal
	// Off turns logging off
	Off
)

// CurrentLogLevel is the current, global log level
var CurrentLogLevel LogLevel = Info

var ll2Str = map[LogLevel]string{
	All:   "ALL",
	Debug: "DEBUG",
	Info:  "INFO",
	Warn:  "WARN",
	Error: "ERROR",
	Fatal: "FATAL",
	Off:   "OFF",
}
var str2LL = map[string]LogLevel{
	"ALL":   All,
	"DEBUG": Debug,
	"INFO":  Info,
	"WARN":  Warn,
	"ERROR": Error,
	"FATAL": Fatal,
	"OFF":   Off,
}

// LogLevelToStr ...
func LogLevelToStr(lvl LogLevel) string {
	return ll2Str[lvl]

}

// StrToLogLevel ...
func StrToLogLevel(str string) (LogLevel, bool) {
	ret, ok := str2LL[str]
	return ret, ok
}

// Println with log level
func Println(level LogLevel, v ...interface{}) {
	if CurrentLogLevel == Off {
		return
	}
	if level >= CurrentLogLevel {
		_, fn, line, _ := runtime.Caller(1)
		log.Println(fmt.Sprintf("%s %s %s:%d %v",
			time.Now().UTC().Format(time.RFC3339),
			LogLevelToStr(level),
			filepath.Base(fn),
			line,
			v))
	}
}

// Fatalln with log level
func Fatalln(v ...interface{}) {
	if CurrentLogLevel == Off {
		return
	}
	log.Fatalln(v...)
}
