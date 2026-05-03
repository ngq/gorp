package log

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
//
// 中文说明：
// - zap 是 framework 的必需依赖核（与 gin/gorm/redis 同级），直接内化在 framework/provider/log；
// - 业务如需深度定制 zap（多 sink、动态 level、自定义 encoder 等），可使用 contrib/log/zap 扩展层。
type SinkConfig struct {
	Driver        string
	Filename      string
	RotatePattern string
	RotateMaxAge  time.Duration
	RotateTime    time.Duration
	MaxSizeMB     int
	MaxBackups    int
	MaxAgeDays    int
	Compress      bool
	LocalTime     bool
}

// zapLogger 是 contract.Logger 的默认 zap 实现。
type zapLogger struct {
	l *zap.Logger
}

// newZapLogger 用默认配置（stdout + console 格式）构造一个最小可用 logger。
func newZapLogger(level, format string) (*zapLogger, error) {
	return newZapLoggerWithSink(level, format, SinkConfig{Driver: "stdout"})
}

// NewDefaultLogger 返回框架默认的 zap logger（info 级别 + console 格式 + stdout）。
//
// 中文说明：
// - 供 framework 内部其他 provider（如 gin provider）在容器尚未就绪时获取兜底 logger；
// - 返回 contract.Logger 接口，调用方无需感知 zap 细节。
func NewDefaultLogger() (contract.Logger, error) {
	return newZapLogger("info", "console")
}

// newZapLoggerWithSink 按 SinkConfig 构造 logger。
func newZapLoggerWithSink(level, format string, sink SinkConfig) (*zapLogger, error) {
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
	return &zapLogger{l: logger}, nil
}

func (z *zapLogger) Debug(msg string, fields ...contract.Field) {
	z.l.Debug(msg, toZapFields(fields)...)
}
func (z *zapLogger) Info(msg string, fields ...contract.Field) {
	z.l.Info(msg, toZapFields(fields)...)
}
func (z *zapLogger) Warn(msg string, fields ...contract.Field) {
	z.l.Warn(msg, toZapFields(fields)...)
}
func (z *zapLogger) Error(msg string, fields ...contract.Field) {
	z.l.Error(msg, toZapFields(fields)...)
}
func (z *zapLogger) With(fields ...contract.Field) contract.Logger {
	return &zapLogger{l: z.l.With(toZapFields(fields)...)}
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

// CloseIfPossible 尝试关闭 logger 资源。
func CloseIfPossible(w any) {
	if c, ok := w.(io.Closer); ok {
		_ = c.Close()
	}
}
