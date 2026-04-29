package log

import (
	logzap "github.com/ngq/gorp/contrib/log/zap"
)

// sinkConfig 是 contrib/log/zap.SinkConfig 的桥接别名。
//
// 中文说明：
// - `framework/provider/log` 只保留最小 bridge；
// - 真实 writer/sink 行为已经迁移到 `contrib/log/zap`。
type sinkConfig = logzap.SinkConfig

// buildWriteSyncer 桥接到 contrib/log/zap 的 writer 构造。
//
// 中文说明：
// - 保留这个桥接函数，是为了让当前核心层测试与少量过渡引用不直接断裂；
// - 新代码应优先直接走 `contrib/log/zap`。
func buildWriteSyncer(sc sinkConfig) (any, error) {
	logger, err := logzap.NewWithSink("info", "console", logzap.SinkConfig(sc))
	if err != nil {
		return nil, err
	}
	return logger, nil
}

// CloseIfPossible 桥接到 contrib/log/zap.CloseIfPossible。
func CloseIfPossible(w any) {
	logzap.CloseIfPossible(w)
}
