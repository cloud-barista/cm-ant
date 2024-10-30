package utils

import (
	"fmt"
	"log"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// LogLevel type to represent different log levels
type LogLevel string

const (
	Info  LogLevel = "INFO"
	Warn  LogLevel = "WARN"
	Error LogLevel = "ERROR"
)

// Utility function for logging messages with color
func Log(level LogLevel, v ...interface{}) {
	var color string
	switch level {
	case Info:
		color = colorGreen
	case Warn:
		color = colorYellow
	case Error:
		color = colorRed
	default:
		color = colorReset
	}
	log.Println(color+fmt.Sprintf("[%s]", level)+colorReset, fmt.Sprint(v...))
}

func Logf(level LogLevel, format string, v ...interface{}) {
	var color string
	switch level {
	case Info:
		color = colorGreen
	case Error:
		color = colorRed
	default:
		color = colorReset
	}
	content := fmt.Sprintf(format, v...)
	log.Println(color+fmt.Sprintf("[%s]", level)+colorReset, content)
}

// Wrapper functions for specific log levels
func LogInfo(v ...interface{}) {
	Log(Info, v...)
}

func LogInfof(format string, v ...interface{}) {
	Logf(Info, format, v...)
}

func LogWarn(v ...interface{}) {
	Log(Warn, v...)
}

func LogWarnf(format string, v ...interface{}) {
	Logf(Warn, format, v...)
}

func LogError(v ...interface{}) {
	Log(Error, v...)
}

func LogErrorf(format string, v ...interface{}) {
	Logf(Error, format, v...)
}
