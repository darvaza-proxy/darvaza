// Package cblog provides a channel based logger.
package cblog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"darvaza.org/slog"
	"darvaza.org/slog/handlers/cblog"
)

var (
	_ slog.Logger = (*Logger)(nil)
)

// LogOutputBuffer is the size of the channel buffer used for logging.
const LogOutputBuffer = 1024

type loggerHandler interface {
	setup(config map[string]any) error
	write(msg cblog.LogMsg)
}

// Logger is the logging object.
type Logger struct {
	*cblog.Logger

	messages <-chan cblog.LogMsg
	outputs  map[string]loggerHandler
}

// New creates a new Logger.
func New() *Logger {
	messages := make(chan cblog.LogMsg, LogOutputBuffer)
	logger, _ := cblog.New(messages)

	l := &Logger{
		Logger:   logger,
		messages: messages,
		outputs:  make(map[string]loggerHandler),
	}
	go l.run(messages)
	return l
}

// SetLogger sets the Logger object output.
func (l *Logger) SetLogger(handlerType string, cfg map[string]any) {
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

	_ = handler.setup(cfg)
	l.outputs[handlerType] = handler
}

func (l *Logger) run(messages <-chan cblog.LogMsg) {
	for msg := range messages {
		for _, handler := range l.outputs {
			handler.write(msg)
		}
	}
}

func stringify(lm cblog.LogMsg) string {
	var out = make([]string, 0, 1)
	var prefix string

	switch lm.Level {
	case slog.Fatal:
		prefix = "[FATAL]"
	case slog.Error:
		prefix = "[ERROR]"
	case slog.Warn:
		prefix = "[WARN]"
	case slog.Info:
		prefix = "[INFO]"
	case slog.Debug:
		prefix = "[DEBUG]"
	default:
		prefix = fmt.Sprintf("[%v]", int(lm.Level))
	}

	out = append(out, prefix)
	if msg := lm.Message; len(msg) > 0 {
		out = append(out, msg)
	}

	return strings.Join(out, " ")
}

type consoleHandler struct {
	logger *log.Logger
}

func newConsoleHandler() loggerHandler {
	return new(consoleHandler)
}

func (h *consoleHandler) setup(_ map[string]any) error {
	h.logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	return nil
}

func (h *consoleHandler) write(lm cblog.LogMsg) {
	fatal := lm.Level == slog.Fatal
	msg := stringify(lm)

	if !fatal {
		h.logger.Println(msg)
	} else {
		h.logger.Fatalln(msg)
	}
}

type fileHandler struct {
	file   string
	logger *log.Logger
}

func newFileHandler() loggerHandler {
	return new(fileHandler)
}

func getConfigString(config map[string]any, key string) (string, bool) {
	if config != nil {
		if v, ok := config[key]; ok {
			if s, ok := v.(string); ok {
				return s, true
			}
		}
	}
	return "", false
}

func (h *fileHandler) setup(config map[string]any) error {
	if file, ok := getConfigString(config, "file"); ok {
		h.file = file

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

func (h *fileHandler) write(lm cblog.LogMsg) {
	if h.logger == nil {
		return
	}

	msg := stringify(lm)

	if lm.Level != slog.Fatal {
		h.logger.Println(msg)
	} else {
		h.logger.Fatalln(msg)
	}
}
