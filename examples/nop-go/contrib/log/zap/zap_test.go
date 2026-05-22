package zap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewWithSinkSingle(t *testing.T) {
	logger, err := NewWithSink("info", "console", SinkConfig{
		Driver:     "single",
		Filename:   "./storage/log/test.log",
		MaxSizeMB:  1,
		MaxBackups: 1,
		MaxAgeDays: 1,
		Compress:   false,
		LocalTime:  true,
	})
	require.NoError(t, err)
	require.NotNil(t, logger)
}

func TestNewWithSinkRotate(t *testing.T) {
	logger, err := NewWithSink("info", "console", SinkConfig{
		Driver:        "rotate",
		Filename:      "./storage/log/test-rotate.log",
		RotatePattern: "./storage/log/test-rotate.log.%Y%m%d%H%M",
		RotateMaxAge:  2 * time.Hour,
		RotateTime:    time.Minute,
	})
	require.NoError(t, err)
	require.NotNil(t, logger)
}
