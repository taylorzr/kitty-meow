package main

import (
	"log"
	"os"
	"path/filepath"
)

var logger *log.Logger

func initLogger() {
	logPath := filepath.Join(meowDir, "meow.log")
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger = log.New(os.Stderr, "", log.LstdFlags)
		return
	}
	logger = log.New(f, "", log.LstdFlags)
}

func logf(format string, args ...any) {
	if logger != nil {
		logger.Printf(format, args...)
	}
}

func logError(context string, err error) {
	if err != nil {
		logf("ERROR %s: %v", context, err)
	}
}
