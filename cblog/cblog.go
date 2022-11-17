// Package cblog provides a channel based logger.
package cblog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// LogOutputBuffer is the size of the channel buffer used for logging.
const LogOutputBuffer = 1024

type logMsg struct {
	msg   string
	fatal bool
}

type loggerHandler interface {
	setup(config map[string]interface{}) error
	write(msg *logMsg)
}

// Logger is the logging object.
type Logger struct {
	messages chan *logMsg
	outputs  map[string]loggerHandler
}

// New creates a new Logger.
func New() *Logger {
	l := &Logger{
		messages: make(chan *logMsg, LogOutputBuffer),
		outputs:  make(map[string]loggerHandler),
	}
	go l.run()
	return l
}

// SetLogger sets the Logger object output.
func (l *Logger) SetLogger(handlerType string, cfg map[string]interface{}) {
	// BUG(karasz): SetLogger should be replaced with SetOutput in order to become stdlib compatible
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
		case msg := <-l.messages:
			for _, handler := range l.outputs {
				handler.write(msg)
			}
		}
	}
}

func (l *Logger) writeMsg(msg string, fatal bool) {
	lm := &logMsg{
		msg:   msg,
		fatal: fatal,
	}
	l.messages <- lm
}

// Debug calls l.writeMsg prefixing the message with [DEBUG]
func (l *Logger) Debug(format string, v ...interface{}) {
	msg := fmt.Sprintf("[DEBUG] "+format, v...)
	l.writeMsg(msg, false)
}

// Info calls l.writeMsg prefixing the message with [INFO]
func (l *Logger) Info(format string, v ...interface{}) {
	msg := fmt.Sprintf("[INFO] "+format, v...)
	l.writeMsg(msg, false)
}

// Notice calls l.writeMsg prefixing the message with [NOTICE]
func (l *Logger) Notice(format string, v ...interface{}) {
	msg := fmt.Sprintf("[NOTICE] "+format, v...)
	l.writeMsg(msg, false)
}

// Warn calls l.writeMsg prefixing the message with [WARN]
func (l *Logger) Warn(format string, v ...interface{}) {
	msg := fmt.Sprintf("[WARN] "+format, v...)
	l.writeMsg(msg, false)
}

// Error calls l.writeMsg prefixing the message with [ERROR]
func (l *Logger) Error(format string, v ...interface{}) {
	msg := fmt.Sprintf("[ERROR] "+format, v...)
	l.writeMsg(msg, false)
}

// Fatal calls l.writeMsg prefixing the message with [FATAL]
func (l *Logger) Fatal(format string, v ...interface{}) {
	msg := fmt.Sprintf("[FATAL] "+format, v...)
	l.writeMsg(msg, true)
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

func (h *consoleHandler) write(lm *logMsg) {
	if !lm.fatal {
		h.logger.Println(lm.msg)
	} else {
		h.logger.Fatalln(lm.msg)
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

func (h *fileHandler) write(lm *logMsg) {
	if h.logger == nil {
		return
	}

	if !lm.fatal {
		h.logger.Println(lm.msg)
	} else {
		h.logger.Fatalln(lm.msg)
	}
}
