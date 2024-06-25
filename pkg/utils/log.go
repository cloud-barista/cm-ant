package utils

import (
	"fmt"
	"log"
)

const (
	colorReset = "\033[0m"
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
)

// Utility function for logging info messages with color
func LogInfo(v ...interface{}) {
	log.Println(colorGreen+"[INFO]"+colorReset, v)
}

func LogInfof(format string, v ...interface{}) {
	content := fmt.Sprintf(format, v...)
	log.Println(colorGreen+"[INFO]"+colorReset, content)
}

// Utility function for logging error messages with color
func LogError(v ...interface{}) {
	log.Println(colorRed+"[ERROR]"+colorReset, v)
}

func LogErrorf(format string, v ...interface{}) {
	content := fmt.Sprintf(format, v...)
	log.Println(colorRed+"[ERROR]"+colorReset, content)
}
