package client

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v5/tracelog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct{}

func (log *Logger) Zap2PgxLogLevel(level zapcore.Level) tracelog.LogLevel {
	switch level {
	case zapcore.DebugLevel:
		return tracelog.LogLevelDebug
	case zapcore.InfoLevel:
		return tracelog.LogLevelWarn
	case zapcore.WarnLevel:
		return tracelog.LogLevelWarn
	case zapcore.ErrorLevel:
		return tracelog.LogLevelError
	case zapcore.DPanicLevel:
		fallthrough
	case zapcore.PanicLevel:
		fallthrough
	case zapcore.FatalLevel:
		fallthrough
	case zapcore.InvalidLevel:
		fallthrough
	default:
		return tracelog.LogLevelError
	}
}

func (log *Logger) Pgx2ZapLogLevel(level tracelog.LogLevel) zapcore.Level {
	switch level {
	case tracelog.LogLevelDebug:
		return zapcore.DebugLevel
	case tracelog.LogLevelInfo:
		return zapcore.InfoLevel
	case tracelog.LogLevelWarn:
		return zapcore.WarnLevel
	case tracelog.LogLevelError:
		return zapcore.ErrorLevel
	}
	return zapcore.ErrorLevel
}

func (log *Logger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]interface{}) {
	l := ctxzap.Extract(ctx)
	l.Log(log.Pgx2ZapLogLevel(level), msg, zap.Reflect("data", data))
}
