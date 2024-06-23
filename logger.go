package prue

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func Log(filePath string, v ...interface{}) {
	logFile, err := getLogFile(filePath)
	if err != nil {
		debugPrintError(fmt.Errorf("failed to get log file: %v", err))
		return
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags|log.Lshortfile)
	logger.Output(2, fmt.Sprintln(v...))
}

func getLogFile(filePath string) (*os.File, error) {
	logDir := filepath.Dir(filePath)

	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return file, nil
}
