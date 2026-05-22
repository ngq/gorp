// Package log_test provides unit tests for log writer build and sink configuration.
//
// 适用场景：
// - 验证 buildWriteSyncer 对不同 SinkConfig 的处理。
// - 确保 single/rolling 等 writer 模式的正确构建。
package log

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestBuildWriteSyncer_Single 验证 single 模式下 writeSyncer 的构建。
//
// 中文说明：
// - single 模式对应单文件写入，buildWriteSyncer 返回有效的 writeSyncer。
func TestBuildWriteSyncer_Single(t *testing.T) {
	ws, err := buildWriteSyncer(SinkConfig{
		Driver:     "single",
		Filename:   "./storage/log/test.log",
		MaxSizeMB:  1,
		MaxBackups: 1,
		MaxAgeDays: 1,
		Compress:   false,
		LocalTime:  true,
	})
	require.NoError(t, err)
	require.NotNil(t, ws)
}

// TestBuildWriteSyncer_Rotate 验证 rotate 模式下 writeSyncer 的构建。
//
// 中文说明：
// - rotate 模式对应日志轮转，buildWriteSyncer 返回有效的 writeSyncer。
func TestBuildWriteSyncer_Rotate(t *testing.T) {
	ws, err := buildWriteSyncer(SinkConfig{
		Driver:        "rotate",
		Filename:      "./storage/log/test-rotate.log",
		RotatePattern: "./storage/log/test-rotate.log.%Y%m%d%H%M",
		RotateMaxAge:  2 * time.Hour,
		RotateTime:    time.Minute,
	})
	require.NoError(t, err)
	require.NotNil(t, ws)
}
