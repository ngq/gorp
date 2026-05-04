package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type gormLogger struct {
	l                         observabilitycontract.Logger
	level                     logger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

func newGormLogger(l observabilitycontract.Logger) logger.Interface {
	return &gormLogger{
		l:                         l,
		level:                     logger.Warn,
		slowThreshold:             200 * time.Millisecond,
		ignoreRecordNotFoundError: true,
	}
}

func (g *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	ng := *g
	ng.level = level
	return &ng
}

func (g *gormLogger) Info(ctx context.Context, msg string, data ...any) {
	if g.level < logger.Info {
		return
	}
	g.l.Info(fmt.Sprintf(msg, data...), observabilitycontract.Field{Key: "source", Value: utils.FileWithLineNum()})
	_ = ctx
}

func (g *gormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if g.level < logger.Warn {
		return
	}
	g.l.Warn(fmt.Sprintf(msg, data...), observabilitycontract.Field{Key: "source", Value: utils.FileWithLineNum()})
	_ = ctx
}

func (g *gormLogger) Error(ctx context.Context, msg string, data ...any) {
	if g.level < logger.Error {
		return
	}
	g.l.Error(fmt.Sprintf(msg, data...), observabilitycontract.Field{Key: "source", Value: utils.FileWithLineNum()})
	_ = ctx
}

func (g *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.level == logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []observabilitycontract.Field{
		{Key: "elapsed_ms", Value: float64(elapsed.Microseconds()) / 1000.0},
		{Key: "rows", Value: rows},
		{Key: "sql", Value: sql},
		{Key: "source", Value: utils.FileWithLineNum()},
	}

	if err != nil && (!g.ignoreRecordNotFoundError || !errors.Is(err, logger.ErrRecordNotFound)) {
		fields = append(fields, observabilitycontract.Field{Key: "error", Value: err.Error()})
		g.l.Error("gorm query error", fields...)
		_ = ctx
		return
	}

	if g.slowThreshold > 0 && elapsed > g.slowThreshold && g.level >= logger.Warn {
		g.l.Warn("gorm slow query", fields...)
		_ = ctx
		return
	}

	if g.level >= logger.Info {
		g.l.Debug("gorm query", fields...)
		_ = ctx
	}
}
