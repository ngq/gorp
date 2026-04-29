package log

import (
	logzap "github.com/ngq/gorp/contrib/log/zap"
)

// ZapLogger 是 contrib/log/zap.Logger 的桥接别名。
//
// 中文说明：
// - `framework/provider/log` 当前只保留最小 bridge；
// - 真实 zap backend 实现已下沉到 `contrib/log/zap`；
// - 保留这个别名是为了让 framework 内剩余引用与过渡测试不直接断裂。
type ZapLogger = logzap.Logger

// NewZapLogger 是 contrib/log/zap.New 的桥接入口。
//
// 中文说明：
// - 核心层不再持有 zap 真实实现；
// - 新代码应优先直接依赖 `contrib/log/zap`。
func NewZapLogger(level, format string) (*ZapLogger, error) {
	return logzap.New(level, format)
}

// NewZapLoggerWithSink 是 contrib/log/zap.NewWithSink 的桥接入口。
//
// 中文说明：
// - 保留这个函数仅用于当前核心层最小桥接；
// - 真实 sink 结构与 writer 行为已经迁移到 contrib 层。
func NewZapLoggerWithSink(level, format string, sink sinkConfig) (*ZapLogger, error) {
	return logzap.NewWithSink(level, format, logzap.SinkConfig(sink))
}
