package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ngq/gorp/framework/contract"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// gormLogger 把 GORM 日志桥接到框架统一 logger。
//
// 中文说明：
// - GORM 自己有一套 logger.Interface，这里负责做适配。
// - 这样 SQL 日志、慢查询、错误日志都能进入统一日志出口。
// - 同时也方便未来在 contract.Logger 层做链路字段、输出格式等统一治理。
type gormLogger struct {
	l                        contract.Logger
	level                    logger.LogLevel
	slowThreshold            time.Duration
	ignoreRecordNotFoundError bool
}

func newGormLogger(l contract.Logger) logger.Interface {
	return &gormLogger{
		l:                        l,
		level:                    logger.Warn,
		slowThreshold:            200 * time.Millisecond,
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
	g.l.Info(fmt.Sprintf(msg, data...), contract.Field{Key: "source", Value: utils.FileWithLineNum()})
	_ = ctx
}

func (g *gormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if g.level < logger.Warn {
		return
	}
	g.l.Warn(fmt.Sprintf(msg, data...), contract.Field{Key: "source", Value: utils.FileWithLineNum()})
	_ = ctx
}

func (g *gormLogger) Error(ctx context.Context, msg string, data ...any) {
	if g.level < logger.Error {
		return
	}
	g.l.Error(fmt.Sprintf(msg, data...), contract.Field{Key: "source", Value: utils.FileWithLineNum()})
	_ = ctx
}

func (g *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if g.level == logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []contract.Field{
		{Key: "elapsed_ms", Value: float64(elapsed.Microseconds()) / 1000.0},
		{Key: "rows", Value: rows},
		{Key: "sql", Value: sql},
		{Key: "source", Value: utils.FileWithLineNum()},
	}

	// error path
	//
	// 中文说明：
	// - SQL 执行报错时优先走 error 分支。
	// - 默认会忽略 record not found，避免正常的“查不到数据”把错误日志刷满。
	if err != nil && (!g.ignoreRecordNotFoundError || !errors.Is(err, logger.ErrRecordNotFound)) {
		fields = append(fields, contract.Field{Key: "error", Value: err.Error()})
		g.l.Error("gorm query error", fields...)
		_ = ctx
		return
	}

	// slow query path
	if g.slowThreshold > 0 && elapsed > g.slowThreshold && g.level >= logger.Warn {
		g.l.Warn("gorm slow query", fields...)
		_ = ctx
		return
	}

	// normal query path
	if g.level >= logger.Info {
		g.l.Debug("gorm query", fields...)
		_ = ctx
	}
}
