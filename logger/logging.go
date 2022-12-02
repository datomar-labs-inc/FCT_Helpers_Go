package lggr

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

type Kind string
type ContextKeyType string

const (
	KindEvent = Kind("event")
	KindState = Kind("state")
)

const ContextKey = ContextKeyType("__lggr.log_wrapper")

type Category string

const (
	CategoryAuthentication     = Category("authentication")
	CategoryConfiguration      = Category("configuration")
	CategoryDatabase           = Category("database")
	CategoryDriver             = Category("driver")
	CategoryFile               = Category("file")
	CategoryHost               = Category("host")
	CategoryIam                = Category("iam")
	CategoryIntrusionDetection = Category("intrusion_detection")
	CategoryMalware            = Category("malware")
	CategoryNetwork            = Category("network")
	CategoryPackage            = Category("package")
	CategoryProcess            = Category("process")
	CategoryRegistry           = Category("registry")
	CategorySession            = Category("session")
	CategoryThreat             = Category("threat")
	CategoryWeb                = Category("web")
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

type LogWrapper struct {
	CallerSkip     int         `json:"caller_skip"`
	DetachedFields []zap.Field `json:"detached_fields"`
	log            *zap.Logger
	ctx            context.Context
}

func New() *LogWrapper {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}

	lg, err := config.Build()
	if err != nil {
		panic(err)
	}

	return &LogWrapper{
		log: lg,
	}
}

func NewFromJSON(parent *LogWrapper, jsonBytes []byte) (*LogWrapper, error) {
	var wrapper LogWrapper

	err := json.Unmarshal(jsonBytes, &wrapper)
	if err != nil {
		return nil, err
	}

	wrapper.log = parent.log.With()

	return &wrapper, nil
}

func NewDev() *LogWrapper {
	lg, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	return &LogWrapper{
		log: lg,
	}
}

func NewTest(t *testing.T) *LogWrapper {
	testLogger := zaptest.NewLogger(t)

	return &LogWrapper{
		log: testLogger,
	}
}

func FromContext(ctx context.Context, action ...string) *LogWrapper {
	if ctx == nil {
		return nil
	}

	if lggr, ok := ctx.Value(ContextKey).(*LogWrapper); ok {

		if len(action) > 0 {
			return lggr.With(zap.String("event.action", action[0]), zap.Namespace("_app_data"))
		}

		lggr.ctx = ctx

		return lggr
	}

	return nil
}

func (log *LogWrapper) GetInternalZapLogger() *zap.Logger {
	return log.log
}

func (log *LogWrapper) AttachToContext(parent context.Context) context.Context {
	if log.log == nil {
		panic("cannot attach logger to context without internal zap logger")
	}

	ctx := context.WithValue(parent, ContextKey, log)
	log.ctx = ctx
	return ctx
}

func (log *LogWrapper) Get(action string) *LogWrapper {
	return log.With(zap.String("event.kind", string(KindEvent))).With(zap.String("event.action", action))
}

// Ctx will attach a context to the logger, it will also attach tracing information
func (log *LogWrapper) Ctx(ctx context.Context) *LogWrapper {
	sc := trace.SpanContextFromContext(ctx)

	if sc.IsValid() {
		if sc.HasSpanID() {
			log.AddFields(zap.String("span.id", sc.SpanID().String()))
		}

		if sc.HasTraceID() {
			log.AddFields(zap.String("trace.id", sc.TraceID().String()))
		}
	}

	return log
}

// StateKind overrides the event.kind field, and sets it to state
func (log *LogWrapper) StateKind() *LogWrapper {
	log.AddFields(zap.String("event.kind", string(KindState)))
	return log
}

func (log *LogWrapper) Category(c Category) *LogWrapper {
	log.AddFields(zap.String("event.category", string(c)))

	return log
}

func (log *LogWrapper) Span(span trace.Span) *LogWrapper {
	if span.SpanContext().HasSpanID() {
		log.AddFields(zap.String("span.id", span.SpanContext().SpanID().String()))
	}

	if span.SpanContext().HasTraceID() {
		log.AddFields(zap.String("trace.id", span.SpanContext().TraceID().String()))
	}

	return log
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *LogWrapper) Info(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 1)).log.Info(msg, fields...)
	for _, hook := range hooks {
		hook.Info(log.ctx, msg, fields...)
	}
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *LogWrapper) Debug(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 0)).log.Debug(msg, fields...)
	for _, hook := range hooks {
		hook.Debug(log.ctx, msg, fields...)
	}
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *LogWrapper) Warn(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 5)).log.Warn(msg, fields...)
	for _, hook := range hooks {
		hook.Warn(log.ctx, msg, fields...)
	}
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *LogWrapper) Error(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 10)).log.Error(msg, fields...)
	for _, hook := range hooks {
		hook.Error(log.ctx, msg, fields...)
	}
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (log *LogWrapper) Panic(msg string, fields ...zap.Field) {
	for _, hook := range hooks {
		hook.Panic(log.ctx, msg, fields...)
	}
	log.With(zap.Int("event.severity", 15)).log.Panic(msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func (log *LogWrapper) Fatal(msg string, fields ...zap.Field) {
	for _, hook := range hooks {
		hook.Fatal(log.ctx, msg, fields...)
	}
	log.With(zap.Int("event.severity", 20)).log.Fatal(msg, fields...)
}

// Critical logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
// Critical will also set a very high event.severity (for elastic)
func (log *LogWrapper) Critical(msg string, fields ...zap.Field) {
	log.With(zap.Int("event.severity", 50)).log.Error(msg, fields...)
	for _, hook := range hooks {
		hook.Critical(log.ctx, msg, fields...)
	}
}

// CriticalPanic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
// CriticalPanic will also set a very high event.severity (for elastic)
func (log *LogWrapper) CriticalPanic(msg string, fields ...zap.Field) {
	for _, hook := range hooks {
		hook.CriticalPanic(log.ctx, msg, fields...)
	}
	log.With(zap.Int("event.severity", 55)).log.Panic(msg, fields...)
}

// CriticalFatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
// CriticalFatal will also set a very high event.severity (for elastic)
func (log *LogWrapper) CriticalFatal(msg string, fields ...zap.Field) {
	for _, hook := range hooks {
		hook.CriticalFatal(log.ctx, msg, fields...)
	}
	log.With(zap.Int("event.severity", 60)).log.Fatal(msg, fields...)
}

func (log *LogWrapper) AddFields(fields ...zap.Field) *LogWrapper {
	log.DetachedFields = append(log.DetachedFields, fields...)
	log.log = log.log.With(fields...)
	return log
}

func (log *LogWrapper) With(fields ...zap.Field) *LogWrapper {
	newLog := &LogWrapper{
		log: log.log,
	}

	newLog = newLog.AddFields(log.DetachedFields...)
	newLog = newLog.AddFields(fields...)
	newLog = newLog.WithCallerSkip(log.CallerSkip)

	return newLog
}

func (log *LogWrapper) WithCallerSkip(n int) *LogWrapper {
	newLog := &LogWrapper{
		CallerSkip: log.CallerSkip,
		log:        log.log,
	}

	newLog = newLog.AddFields(log.DetachedFields...)
	newLog.log = newLog.log.WithOptions(zap.AddCallerSkip(n))
	newLog.CallerSkip += n

	return newLog
}

func (log *LogWrapper) Sync() error {
	return log.log.Sync()
}
