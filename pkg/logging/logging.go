package logging

import (
	"github.com/verte-zerg/gession/pkg/assert"
	"log/slog"
	"os"
	"path"
	"sync"

	"github.com/adrg/xdg"
)

const (
	fileOpenFlags = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	fileOpenMode  = 0644
	logDirMode    = 0755
)

var (
	logger *Logger
	once   sync.Once
)

func GetInstance() *Logger {
	once.Do(
		func() {
			logger = newLogger()
		},
	)

	return logger
}

func ensureDirExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, logDirMode)
		assert.Assert(err == nil, "could not create directory: %s", dir)
	}
}

type Logger struct {
	logger *slog.Logger
}

func newLogger() *Logger {
	opts := slog.HandlerOptions{Level: slog.LevelDebug}

	logDir := path.Join(xdg.StateHome, "gession")
	ensureDirExists(logDir)

	logFileName := path.Join(logDir, "gession.log")
	logFile, err := os.OpenFile(logFileName, fileOpenFlags, fileOpenMode)
	assert.Assert(err == nil, "could not open log file")

	l := slog.New(slog.NewTextHandler(logFile, &opts))

	return &Logger{logger: l}
}

func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

func (l *Logger) WithGroup(group string) *Logger {
	return &Logger{logger: l.logger.WithGroup(group)}
}
