package internal

import (
	"log"
	"os"
)

var logger = log.New(os.Stdout, "", log.Ltime)

func Log(format string, v ...any) {
	logger.Printf(format+"\n", v...)
}

func LogDebug(format string, v ...any) {
	logger.Printf(format+"\n", v...)
}

func LogFatal(format string, v ...any) {
	logger.Fatalf(format+"\n", v...)
}
