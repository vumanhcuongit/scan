package logger

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

type GormLogger struct {
	LogLevel      gormlogger.LogLevel
	SlowThreshold time.Duration
}

func NewGormLogger() GormLogger {
	return GormLogger{
		LogLevel:      gormlogger.Warn,
		SlowThreshold: 100 * time.Millisecond,
	}
}

func (l GormLogger) SetAsDefault() {
	gormlogger.Default = l
}

func (l GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return GormLogger{
		SlowThreshold: l.SlowThreshold,
		LogLevel:      level,
	}
}

func (l GormLogger) Info(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Info {
		return
	}
	l.logger(ctx).Sugar().Infof(str, args...)
}

func (l GormLogger) Warn(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Warn {
		return
	}
	l.logger(ctx).Sugar().Warnf(str, args...)
}

func (l GormLogger) Error(ctx context.Context, str string, args ...interface{}) {
	if l.LogLevel < gormlogger.Error {
		return
	}
	l.logger(ctx).Sugar().Errorf(str, args...)
}

func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= 0 {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormlogger.Error:
		sql, rows := fc()
		l.logger(ctx).Error(fmt.Sprintf("sql_query: %s", sql), zap.Error(err), zap.String("elapsed", elapsed.String()), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		l.logger(ctx).Warn(fmt.Sprintf("sql_query: %s", sql), zap.String("elapsed", elapsed.String()), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.LogLevel >= gormlogger.Info:
		sql, rows := fc()
		l.logger(ctx).Info(fmt.Sprintf("sql_query: %s", sql), zap.String("elapsed", elapsed.String()), zap.Int64("rows", rows), zap.String("sql", sql))
	}
}

func (l GormLogger) logger(ctx context.Context) *zap.Logger {
	logger := ctxzap.Extract(ctx)
	for i := 2; i < 15; i++ {
		_, file, _, ok := runtime.Caller(i)
		switch {
		case !ok:
		case strings.HasSuffix(file, "_test.go"):
		case strings.Contains(file, filepath.Join("gorm.io", "gorm")):
		default:
			return logger.WithOptions(zap.AddCallerSkip(i))
		}
	}

	return logger
}
