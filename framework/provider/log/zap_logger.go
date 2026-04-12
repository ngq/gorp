package log

import (
	"fmt"

	"github.com/ngq/gorp/framework/contract"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger 是 contract.Logger 的 zap 实现。
//
// 中文说明：
// - 对外仍暴露框架自己的 Logger 接口，避免业务层被具体日志库绑死。
// - 内部真实执行者是 zap.Logger，负责高性能结构化日志输出。
type ZapLogger struct {
	l *zap.Logger
}

func NewZapLogger(level, format string) (*ZapLogger, error) {
	return NewZapLoggerWithSink(level, format, sinkConfig{Driver: "stdout"})
}

// NewZapLoggerWithSink 根据级别、编码格式和输出端配置创建 logger。
//
// 中文说明：
// - level 决定日志过滤级别。
// - format 决定编码格式，目前支持 console/json。
// - sink 决定最终写到哪里，例如 stdout、单文件或按时间滚动文件。
func NewZapLoggerWithSink(level, format string, sink sinkConfig) (*ZapLogger, error) {
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

	// writer
	if err := ensureDir(sink.Filename); err != nil {
		return nil, err
	}
	ws, err := buildWriteSyncer(sink)
	if err != nil {
		return nil, err
	}

	core := zapcore.NewCore(enc, ws, lvl)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return &ZapLogger{l: logger}, nil
}

func (z *ZapLogger) Debug(msg string, fields ...contract.Field) {
	z.l.Debug(msg, toZapFields(fields)...)
}

func (z *ZapLogger) Info(msg string, fields ...contract.Field) {
	z.l.Info(msg, toZapFields(fields)...)
}

func (z *ZapLogger) Warn(msg string, fields ...contract.Field) {
	z.l.Warn(msg, toZapFields(fields)...)
}

func (z *ZapLogger) Error(msg string, fields ...contract.Field) {
	z.l.Error(msg, toZapFields(fields)...)
}

func (z *ZapLogger) With(fields ...contract.Field) contract.Logger {
	return &ZapLogger{l: z.l.With(toZapFields(fields)...)}
}

func toZapFields(fields []contract.Field) []zap.Field {
	out := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		out = append(out, zap.Any(f.Key, f.Value))
	}
	return out
}
