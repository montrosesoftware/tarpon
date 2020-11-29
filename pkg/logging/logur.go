package logging

import (
	"github.com/sirupsen/logrus"
	logrusadapter "logur.dev/adapter/logrus"
	"logur.dev/logur"
)

type LogurLogger struct {
	logur.Logger
}

func NewLogurLogger(logger logur.Logger) *LogurLogger {
	return &LogurLogger{
		Logger: logger,
	}
}

func NewLogrusLogger() *LogurLogger {
	logrus := logrus.New()
	return NewLogurLogger(logrusadapter.New(logrus))
}

func (l *LogurLogger) WithFields(fields map[string]interface{}) Logger {
	return &LogurLogger{Logger: logur.WithFields(l.Logger, fields)}
}
