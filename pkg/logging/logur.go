package logging

import (
	"github.com/montrosesoftware/tarpon/pkg/config"
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

func NewLogrusLogger(config *config.Logging) *LogurLogger {
	logger := logrus.New()

	if level, err := logrus.ParseLevel(config.Level); err != nil {
		logger.Errorf("can't set log level to %s: %v", config.Level, err)
	} else {
		logger.SetLevel(level)
	}

	return NewLogurLogger(logrusadapter.New(logger))
}

func (l *LogurLogger) WithFields(fields map[string]interface{}) Logger {
	return &LogurLogger{Logger: logur.WithFields(l.Logger, fields)}
}
