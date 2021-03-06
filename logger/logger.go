package logger

import (
	"context"
	"log"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	CtxUserAgentKey     = "User-Agent"
	CtxCorrelationIDKey = "X-Correlation-ID"
)

type (
	Config struct {
		EnableStackTrace bool `yaml:"enableStackTrace"`
	}
	Field struct {
		Key   string `json:"key"`
		Value any    `json:"value"`
	}
)

var (
	logger *zap.Logger
	once   sync.Once
)

func Init(config *Config) {
	var (
		zapConfig = zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		encoder = zapcore.NewJSONEncoder(zapConfig)
		level   = zapcore.InfoLevel
		core    = zapcore.NewCore(encoder, os.Stdout, level)
		opts    = []zap.Option{zap.AddCaller(), zap.AddCallerSkip(1)}
	)

	if config.EnableStackTrace {
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	}

	once.Do(func() {
		logger = zap.New(core, opts...)
	})
}

func getUserAgentFromContext(ctx context.Context) (userAgent string) {
	val := ctx.Value(CtxUserAgentKey)
	if val != nil {
		userAgent = val.(string)
	}
	return
}

func getCorrelationIDFromContext(ctx context.Context) (id string) {
	val := ctx.Value(CtxCorrelationIDKey)
	if val != nil {
		id = val.(string)
	}
	return
}

func transformFields(ctx context.Context, fields []Field) []zap.Field {
	if logger == nil {
		log.Panic("error: logger uninitialized, exec logger.Init() first.")
	}

	var (
		zapFields     = make([]zap.Field, 0, 2)
		zapMeta       = make(map[string]any, 2)
		userAgent     = getUserAgentFromContext(ctx)
		correlationID = getCorrelationIDFromContext(ctx)
		zapData       = make(map[string]any, len(fields))
	)

	if userAgent != "" {
		zapMeta[CtxUserAgentKey] = userAgent
	}

	if correlationID != "" {
		zapMeta[CtxCorrelationIDKey] = correlationID
	}

	if len(zapMeta) > 0 {
		zapFields = append(zapFields, zap.Any("meta", zapMeta))
	}

	for _, f := range fields {
		zapData[f.Key] = f.Value
	}

	if len(zapData) > 0 {
		zapFields = append(zapFields, zap.Any("data", zapData))
	}

	return zapFields
}

func Error(ctx context.Context, msg string, fields ...Field) {
	logger.Error(msg, transformFields(ctx, fields)...)
}

func Warn(ctx context.Context, msg string, fields ...Field) {
	logger.Warn(msg, transformFields(ctx, fields)...)
}

func Info(ctx context.Context, msg string, fields ...Field) {
	logger.Info(msg, transformFields(ctx, fields)...)
}

func Debug(ctx context.Context, msg string, fields ...Field) {
	logger.Info(msg, transformFields(ctx, fields)...)
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func Sync() error {
	return logger.Sync()
}
