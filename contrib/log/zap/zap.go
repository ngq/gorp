package zap

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ngq/gorp/framework/contract"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// SinkConfig 描述日志输出端配置。
type SinkConfig struct {
	Path string
	Driver string
	Filename string
	RotatePattern string
	RotateMaxAge time.Duration
	RotateTime time.Duration
	MaxSizeMB int
	MaxBackups int
	MaxAgeDays int
	Compress bool
	LocalTime bool
}

// Logger 是 contract.Logger 的 zap 实现。
type Logger struct {
	l *zap.Logger
}

func New(level, format string) (*Logger, error) {
	return NewWithSink(level, format, SinkConfig{Driver: "stdout"})
}

func NewWithSink(level, format string, sink SinkConfig) (*Logger, error) {
	lvl := zapcore.InfoLevel
	if err := lvl.Set(level); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", level, err)
	}

	encCfg := zap.NewProductionEncoderConfig()
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var enc zapcore.Encoder
	switch format {
	case "json":
		enc = zapcore.NewJSONEncoder(encCfg)
	case "console", "text", "pretty", "":
		enc = zapcore.NewConsoleEncoder(encCfg)
	default:
		return nil, fmt.Errorf("unknown log format: %s", format)
	}

	if err := ensureDir(sink.Filename); err != nil {
		return nil, err
	}
	ws, err := buildWriteSyncer(sink)
	if err != nil {
		return nil, err
	}

	core := zapcore.NewCore(enc, ws, lvl)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return &Logger{l: logger}, nil
}

func (z *Logger) Debug(msg string, fields ...contract.Field) { z.l.Debug(msg, toZapFields(fields)...) }
func (z *Logger) Info(msg string, fields ...contract.Field)  { z.l.Info(msg, toZapFields(fields)...) }
func (z *Logger) Warn(msg string, fields ...contract.Field)  { z.l.Warn(msg, toZapFields(fields)...) }
func (z *Logger) Error(msg string, fields ...contract.Field) { z.l.Error(msg, toZapFields(fields)...) }
func (z *Logger) With(fields ...contract.Field) contract.Logger {
	return &Logger{l: z.l.With(toZapFields(fields)...)}
}

func toZapFields(fields []contract.Field) []zap.Field {
	out := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		out = append(out, zap.Any(f.Key, f.Value))
	}
	return out
}

func buildWriteSyncer(sc SinkConfig) (zapcore.WriteSyncer, error) {
	switch sc.Driver {
	case "", "stdout":
		return zapcore.AddSync(os.Stdout), nil
	case "single":
		w := &lumberjack.Logger{
			Filename:   sc.Filename,
			MaxSize:    sc.MaxSizeMB,
			MaxBackups: sc.MaxBackups,
			MaxAge:     sc.MaxAgeDays,
			Compress:   sc.Compress,
			LocalTime:  sc.LocalTime,
		}
		return zapcore.AddSync(w), nil
	case "rotate":
		w, err := rotatelogs.New(
			sc.RotatePattern,
			rotatelogs.WithLinkName(sc.Filename),
			rotatelogs.WithMaxAge(sc.RotateMaxAge),
			rotatelogs.WithRotationTime(sc.RotateTime),
		)
		if err != nil {
			return nil, err
		}
		return zapcore.AddSync(w), nil
	default:
		return nil, os.ErrInvalid
	}
}

func ensureDir(p string) error {
	if p == "" {
		return nil
	}
	return os.MkdirAll(filepath.Dir(p), 0o755)
}

func CloseIfPossible(w any) {
	if c, ok := w.(io.Closer); ok {
		_ = c.Close()
	}
}
