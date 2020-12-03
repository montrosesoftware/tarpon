package logging

type Fields map[string]interface{}

type Logger interface {
	Trace(msg string, fields ...map[string]interface{})
	Debug(msg string, fields ...map[string]interface{})
	Info(msg string, fields ...map[string]interface{})
	Warn(msg string, fields ...map[string]interface{})
	Error(msg string, fields ...map[string]interface{})

	IsDebug() bool

	WithFields(fields map[string]interface{}) Logger
}

// NoopLogger is a logger that discards every log event.
type NoopLogger struct{}

func (NoopLogger) Trace(_ string, _ ...map[string]interface{}) {}
func (NoopLogger) Debug(_ string, _ ...map[string]interface{}) {}
func (NoopLogger) Info(_ string, _ ...map[string]interface{})  {}
func (NoopLogger) Warn(_ string, _ ...map[string]interface{})  {}
func (NoopLogger) Error(_ string, _ ...map[string]interface{}) {}
func (NoopLogger) IsDebug() bool                               { return false }

func (n NoopLogger) WithFields(_ map[string]interface{}) Logger { return n }
