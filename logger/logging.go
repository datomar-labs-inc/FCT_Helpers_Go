package lggr

import "go.uber.org/zap"

var logger *zap.Logger

type Kind string

const (
	KindEvent = Kind("event")
	KindState = Kind("state")
)

type Category string

const (
	CategorySomething = Category("something")
)

type ElasticLoggingFields struct {

	// https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-kind
	// Will most likely be event
	Kind Kind

	// https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-category
	Category Category

	// https://www.elastic.co/guide/en/ecs/current/ecs-event.html#field-event-action
	Action string
}

func init() {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	lg, err := config.Build()
	if err != nil {
		panic(err)
	}

	logger = lg
}

type LogWrapper struct {
	*zap.Logger
}

// GetLogger returns a new logger wrapping the zap logger with a default event.kind of "event"
func GetLogger() *LogWrapper {
	return &LogWrapper{
		Logger: logger.With(zap.String("event.kind", string(KindEvent))),
	}
}

// StateKind overrides the event.kind field, and sets it to state
func (log *LogWrapper) StateKind() *LogWrapper {
	log.Logger = log.Logger.With(zap.String("event.kind", string(KindState)))
	return log
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *LogWrapper) Info(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 0)).Info(msg, fields...)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *LogWrapper) Debug(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 0)).Debug(msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *LogWrapper) Warn(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 5)).Warn(msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *LogWrapper) Error(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 10)).Error(msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (log *LogWrapper) Panic(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 15)).Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func (log *LogWrapper) Fatal(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 20)).Fatal(msg, fields...)
}

// Critical logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
// Critical will also set a very high event.severity (for elastic)
func (log *LogWrapper) Critical(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 50)).Error(msg, fields...)
}

// CriticalPanic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
// CriticalPanic will also set a very high event.severity (for elastic)
func (log *LogWrapper) CriticalPanic(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 55)).Panic(msg, fields...)
}

// CriticalFatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
// CriticalFatal will also set a very high event.severity (for elastic)
func (log *LogWrapper) CriticalFatal(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 60)).Fatal(msg, fields...)
}
