package logger

import (
	"bytes"
	"io"
	"log"
	"os"
	"sync"
)

// Logger is the global logger instance.
var Logger *log.Logger

// buffer holds the in-memory log messages.
var buffer bytes.Buffer

// bufferMutex protects access to the buffer.
var bufferMutex sync.Mutex

// InitLogger initializes the logger to write to both a file and an in-memory buffer.
func InitLogger() error {
	// Open or create the log file.
	f, err := os.OpenFile("aztui.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	// Create a MultiWriter to write to both the file and the buffer.
	mw := io.MultiWriter(f, &buffer)

	// Initialize the logger.
	Logger = log.New(mw, "", log.LstdFlags)
	Logger.Println("Starting log")

	return nil
}

// Println logs a message using the global Logger.
func Println(v ...interface{}) {
	Logger.Println(v...)
}

// GetLogs returns the current contents of the in-memory log buffer.
func GetLogs() string {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()
	return buffer.String()
}

// ClearLogs clears the in-memory log buffer.
func ClearLogs() {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()
	buffer.Reset()
}
