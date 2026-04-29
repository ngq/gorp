package zap

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRejectsInvalidLevel(t *testing.T) {
	_, err := New("not-a-level", "console")
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid log level")
}

func TestNewRejectsUnknownFormat(t *testing.T) {
	_, err := New("info", "xml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown log format")
}

func TestNewWithSinkRejectsUnknownDriver(t *testing.T) {
	_, err := NewWithSink("info", "console", SinkConfig{Driver: "mystery"})
	require.Error(t, err)
	require.ErrorIs(t, err, os.ErrInvalid)
}

func TestToZapFieldsAndWithLogger(t *testing.T) {
	logger, err := New("info", "console")
	require.NoError(t, err)
	child := logger.With()
	require.NotNil(t, child)
	require.NotPanics(t, func() {
		logger.Info("hello")
		logger.Warn("warn")
		logger.Error("error")
	})
}
