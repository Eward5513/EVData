package common

import (
	"fmt"
	"io"
	"log"
	"os"
)

var logFile *os.File

func CreatLogFile(filename string) {
	var err error
	logFile, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("error opening file:" + err.Error())
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

func CloseLogFile() {
	_ = logFile.Close()
}

type LogLevel = int

const (
	DEBUG LogLevel = iota
	INFO  LogLevel = iota
	WARN  LogLevel = iota
	ERROR LogLevel = iota
)

var logLevel = INFO // 设置日志级别

func SetLogLevel(level int) {
	logLevel = level
}

func DebugLog(msg ...any) {
	if logLevel <= DEBUG {
		log.Println("[DEBUG]", msg)
	}
}

func InfoLog(msg ...any) {
	if logLevel <= INFO {
		log.Println("[INFO]", msg)
	}
}

func WarnLog(msg ...any) {
	if logLevel <= WARN {
		log.Println("[WARN]", msg)
	}
}

func ErrorLog(msg ...any) {
	if logLevel <= ERROR {
		log.Println("[ERROR]", msg)
	}
}

func FatalLog(msg ...any) {
	log.Fatalln("[FATAL]", msg)
}
