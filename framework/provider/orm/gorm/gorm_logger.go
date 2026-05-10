// Package gorm provides a custom GORM logger implementation.
// The logger supports log level control, slow query detection, and error recording.
// Default settings:
//
// 本文件提供自定义的 GORM 日志实现，集成 gorp 框架日志系统。
// 支持日志级别控制、慢查询检测、错误记录。
// 默认设置：
//   - LogLevel: Warn
//   - SlowThreshold: 200ms
//   - IgnoreRecordNotFoundError: true (ErrRecordNotFound is not logged as error)
//
// 默认设置说明：ErrRecordNotFound 不作为错误记录，避免不必要的日志噪音。
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

// gormLogger implements logger.Interface for GORM using gorp's Logger.
//
// gormLogger 实现 GORM 的 logger.Interface，使用 gorp 框架的 Logger。
type gormLogger struct {
	l                         observabilitycontract.Logger // l is the framework logger.
	                                                       //
	                                                        // l 框架日志实例。
	level                     logger.LogLevel              // level is the current log level.
	                                                       //
	                                                        // level 当前日志级别。
	slowThreshold             time.Duration                // slowThreshold is slow query threshold.
	                                                       //
	                                                        // slowThreshold 慢查询阈值。
	ignoreRecordNotFoundError bool                         // ignoreRecordNotFoundError controls ErrRecordNotFound handling.
	                                                       //
	                                                        // ignoreRecordNotFoundError 是否忽略 ErrRecordNotFound。
}

// newGormLogger creates a new GORM logger adapter.
//
// newGormLogger 创建新的 GORM 日志适配器。
func newGormLogger(l observabilitycontract.Logger) logger.Interface {
	return &gormLogger{
		l:                         l,
		level:                     logger.Warn,
		slowThreshold:             200 * time.Millisecond,
		ignoreRecordNotFoundError: true,
	}
}

// LogMode returns a new logger with the specified log level.
//
// LogMode 返回指定日志级别的新日志器实例。
func (g *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	ng := *g
	ng.level = level
	return &ng
}

// Info logs an info-level message.
//
// Info 记录信息级别的日志消息。
func (g *gormLogger) Info(ctx context.Context, msg string, data ...any) {
	if g.level < logger.Info {
		return
	}
	g.l.Info(fmt.Sprintf(msg, data...), observabilitycontract.Field{Key: "source", Value: utils.FileWithLineNum()})
	_ = ctx
}

// Warn logs a warn-level message.
//
// Warn 记录警告级别的日志消息。
func (g *gormLogger) Warn(ctx context.Context, msg string, data ...any) {
	if g.level < logger.Warn {
		return
	}
	g.l.Warn(fmt.Sprintf(msg, data...), observabilitycontract.Field{Key: "source", Value: utils.FileWithLineNum()})
	_ = ctx
}

// Error logs an error-level message.
//
// Error 记录错误级别的日志消息。
func (g *gormLogger) Error(ctx context.Context, msg string, data ...any) {
	if g.level < logger.Error {
		return
	}
	g.l.Error(fmt.Sprintf(msg, data...), observabilitycontract.Field{Key: "source", Value: utils.FileWithLineNum()})
	_ = ctx
}

// Trace logs SQL query execution details with elapsed time and error info.
// Core logic: Check log level, then route to Error/Warn/Debug based on conditions.
//
// Trace 记录 SQL 查询执行详情，包括耗时和错误信息。
// 核心逻辑：检查日志级别，然后根据条件路由到 Error/Warn/Debug。
//   - Error: query failed (except ErrRecordNotFound if ignored)
//   - Warn: slow query exceeds threshold
//   - Debug: normal query
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