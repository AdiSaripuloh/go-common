package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	CtxUserAgentKey     = "User-Agent"
	CtxCorrelationIDKey = "X-Correlation-ID"
)

type Field struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

var logger *zap.Logger

func Init(stackTraceEnabled bool) {
	var (
		config = zapcore.EncoderConfig{
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
		encoder = zapcore.NewJSONEncoder(config)
		level   = zapcore.InfoLevel
		core    = zapcore.NewCore(encoder, os.Stdout, level)
		opts    = []zap.Option{zap.AddCaller(), zap.AddCallerSkip(1)}
	)

	if stackTraceEnabled {
		opts = append(opts, zap.AddStacktrace(zap.ErrorLevel))
	}

	logger = zap.New(core, opts...)
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
	var (
		zapFields     = make([]zap.Field, 0, 2)
		zapMeta       = make([]Field, 0, 2)
		userAgent     = getUserAgentFromContext(ctx)
		correlationID = getCorrelationIDFromContext(ctx)
		zapData       = make([]Field, 0, len(fields))
	)

	if userAgent != "" {
		zapMeta = append(zapMeta, Field{CtxUserAgentKey, userAgent})
	}

	if correlationID != "" {
		zapMeta = append(zapMeta, Field{CtxCorrelationIDKey, correlationID})
	}

	if len(zapMeta) > 0 {
		zapFields = append(zapFields, zap.Any("meta", zapMeta))
	}

	for _, f := range fields {
		zapData = append(zapData, f)
	}

	if len(zapData) > 0 {
		zapFields = append(zapFields, zap.Any("data", zapData))
	}

	return zapFields
}

func Error(ctx context.Context, msg string, fields ...Field) {
	defer logger.Sync()
	logger.Error(msg, transformFields(ctx, fields)...)
}

func Warn(ctx context.Context, msg string, fields ...Field) {
	defer logger.Sync()
	logger.Warn(msg, transformFields(ctx, fields)...)
}

func Info(ctx context.Context, msg string, fields ...Field) {
	defer logger.Sync()
	logger.Info(msg, transformFields(ctx, fields)...)
}

func Debug(ctx context.Context, msg string, fields ...Field) {
	defer logger.Sync()
	logger.Info(msg, transformFields(ctx, fields)...)
}
