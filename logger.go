package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

const LOG_OUTPUT_BUFFER = 1024

type logMesg struct {
	Mesg  string
	Fatal bool
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

func (l *GnoccoLogger) writeMesg(mesg string, fatal bool) {
	lm := &logMesg{
		Mesg:  mesg,
		Fatal: fatal,
	}
	l.mesgs <- lm
}

func (l *GnoccoLogger) Debug(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[DEBUG] "+format, v...)
	l.writeMesg(mesg, false)
}

func (l *GnoccoLogger) Info(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[INFO] "+format, v...)
	l.writeMesg(mesg, false)
}

func (l *GnoccoLogger) Notice(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[NOTICE] "+format, v...)
	l.writeMesg(mesg, false)
}

func (l *GnoccoLogger) Warn(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[WARN] "+format, v...)
	l.writeMesg(mesg, false)
}

func (l *GnoccoLogger) Error(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[ERROR] "+format, v...)
	l.writeMesg(mesg, false)
}

func (l *GnoccoLogger) Fatal(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[FATAL] "+format, v...)
	l.writeMesg(mesg, true)
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
	if !lm.Fatal {
		h.logger.Println(lm.Mesg)
	} else {
		h.logger.Fatalln(lm.Mesg)
	}
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
		if _, err := os.Stat(h.file); os.IsNotExist(err) {
			if _, err := os.Create(h.file); err != nil {
				if errP := os.MkdirAll(filepath.Dir(h.file), 0755); errP != nil {
					return errP
				}
			}
		}

		usr, errU := user.Lookup(Config.User)
		if errU != nil {
			return errU
		}
		uid, _ := strconv.Atoi(usr.Uid)
		gid, _ := strconv.Atoi(usr.Gid)
		output, err := os.OpenFile(h.file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		if errChown := os.Chown(h.file, uid, gid); errChown != nil {
			fmt.Println(errChown)
		}

		h.logger = log.New(output, "", log.Ldate|log.Ltime)
	}

	return nil
}

func (h *FileHandler) Write(lm *logMesg) {
	if h.logger == nil {
		return
	}

	if !lm.Fatal {
		h.logger.Println(lm.Mesg)
	} else {
		h.logger.Fatalln(lm.Mesg)
	}
}
