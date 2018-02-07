package cblog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const LOG_OUTPUT_BUFFER = 1024

type logMesg struct {
	mesg  string
	fatal bool
}

type loggerHandler interface {
	setup(config map[string]interface{}) error
	write(mesg *logMesg)
}

type Logger struct {
	messages chan *logMesg
	outputs  map[string]loggerHandler
}

func New() *Logger {
	l := &Logger{
		messages: make(chan *logMesg, LOG_OUTPUT_BUFFER),
		outputs:  make(map[string]loggerHandler),
	}
	go l.run()
	return l
}

func (l *Logger) SetLogger(handlerType string, cfg map[string]interface{}) {
	var handler loggerHandler
	switch handlerType {
	case "console":
		handler = newConsoleHandler()
	case "file":
		handler = newFileHandler()
	default:
		panic("Unknown log handler.")
	}

	handler.setup(cfg)
	l.outputs[handlerType] = handler
}

func (l *Logger) run() {
	for {
		select {
		case mesg := <-l.messages:
			for _, handler := range l.outputs {
				handler.write(mesg)
			}
		}
	}
}

func (l *Logger) writeMesg(mesg string, fatal bool) {
	lm := &logMesg{
		mesg:  mesg,
		fatal: fatal,
	}
	l.messages <- lm
}

func (l *Logger) Debug(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[DEBUG] "+format, v...)
	l.writeMesg(mesg, false)
}

func (l *Logger) Info(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[INFO] "+format, v...)
	l.writeMesg(mesg, false)
}

func (l *Logger) Notice(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[NOTICE] "+format, v...)
	l.writeMesg(mesg, false)
}

func (l *Logger) Warn(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[WARN] "+format, v...)
	l.writeMesg(mesg, false)
}

func (l *Logger) Error(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[ERROR] "+format, v...)
	l.writeMesg(mesg, false)
}

func (l *Logger) Fatal(format string, v ...interface{}) {
	mesg := fmt.Sprintf("[FATAL] "+format, v...)
	l.writeMesg(mesg, true)
}

type consoleHandler struct {
	logger *log.Logger
}

func newConsoleHandler() loggerHandler {
	return new(consoleHandler)
}

func (h *consoleHandler) setup(cfg map[string]interface{}) error {
	h.logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	return nil
}

func (h *consoleHandler) write(lm *logMesg) {
	if !lm.fatal {
		h.logger.Println(lm.mesg)
	} else {
		h.logger.Fatalln(lm.mesg)
	}
}

type fileHandler struct {
	file   string
	logger *log.Logger
}

func newFileHandler() loggerHandler {
	return new(fileHandler)
}

func (h *fileHandler) setup(config map[string]interface{}) error {
	if file, ok := config["file"]; ok {
		h.file = file.(string)
		if _, err := os.Stat(h.file); os.IsNotExist(err) {
			_ = os.MkdirAll(filepath.Dir(h.file), 0755)
			if _, err := os.Create(h.file); err != nil {
				return err
			}
		}

		output, _ := os.Create(h.file)
		h.logger = log.New(output, "", log.Ldate|log.Ltime)
	}

	return nil
}

func (h *fileHandler) write(lm *logMesg) {
	if h.logger == nil {
		return
	}

	if !lm.fatal {
		h.logger.Println(lm.mesg)
	} else {
		h.logger.Fatalln(lm.mesg)
	}
}
