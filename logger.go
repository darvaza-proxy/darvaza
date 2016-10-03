package main

import (
	"fmt"
	"log"
	"os"
)

const LOG_OUTPUT_BUFFER = 1024

type logMesg struct {
	Mesg string
}

type LoggerHandler interface {
	Setup(config map[string]interface{}) error
	Write(mesg *logMesg)
}

type GnoccoLogger struct {
	mesgs   chan *logMesg
	outputs map[string]LoggerHandler
}

func NewLogger() *GnoccoLogger {
	logger := &GnoccoLogger{
		mesgs:   make(chan *logMesg, LOG_OUTPUT_BUFFER),
		outputs: make(map[string]LoggerHandler),
	}
	go logger.Run()
	return logger
}

func (l *GnoccoLogger) SetLogger(handlerType string, config map[string]interface{}) {
	var handler LoggerHandler
	switch handlerType {
	case "console":
		handler = NewConsoleHandler()
	case "file":
		handler = NewFileHandler()
	default:
		panic("Unknown log handler.")
	}

	handler.Setup(config)
	l.outputs[handlerType] = handler
}

func (l *GnoccoLogger) Run() {
	for {
		select {
		case mesg := <-l.mesgs:
			for _, handler := range l.outputs {
				handler.Write(mesg)
			}
		}
	}
}

func (l *GnoccoLogger) writeMesg(mesg string) {

	lm := &logMesg{
		Mesg: mesg,
	}

	l.mesgs <- lm
}

func (l *GnoccoLogger) Debug(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[DEBUG] "+format, v...)
	l.writeMesg(mesg)
}

func (l *GnoccoLogger) Info(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[INFO] "+format, v...)
	l.writeMesg(mesg)
}

func (l *GnoccoLogger) Notice(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[NOTICE] "+format, v...)
	l.writeMesg(mesg)
}

func (l *GnoccoLogger) Warn(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[WARN] "+format, v...)
	l.writeMesg(mesg)
}

func (l *GnoccoLogger) Error(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[ERROR] "+format, v...)
	l.writeMesg(mesg)
}

type ConsoleHandler struct {
	logger *log.Logger
}

func NewConsoleHandler() LoggerHandler {
	return new(ConsoleHandler)
}

func (h *ConsoleHandler) Setup(config map[string]interface{}) error {
	h.logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	return nil

}

func (h *ConsoleHandler) Write(lm *logMesg) {
	h.logger.Println(lm.Mesg)
}

type FileHandler struct {
	file   string
	logger *log.Logger
}

func NewFileHandler() LoggerHandler {
	return new(FileHandler)
}

func (h *FileHandler) Setup(config map[string]interface{}) error {
	if file, ok := config["file"]; ok {
		h.file = file.(string)
		output, err := os.Create(h.file)
		if err != nil {
			return err
		}

		h.logger = log.New(output, "", log.Ldate|log.Ltime)
	}

	return nil
}

func (h *FileHandler) Write(lm *logMesg) {
	if h.logger == nil {
		return
	}

	h.logger.Println(lm.Mesg)
}
