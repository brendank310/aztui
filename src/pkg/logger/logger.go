package logger

import (
	"log"
	"os"
)

var Logger *log.Logger

func InitLogger() error {
	f, err := os.OpenFile("aztui.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	Logger = log.New(f, "", log.LstdFlags)
	Logger.Println("Starting log")
	return nil
}

func Println(v ...interface{}) {
	Logger.Println(v...)
}
