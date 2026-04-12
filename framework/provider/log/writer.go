package log

import (
	"io"
	"os"
	"path/filepath"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// sinkConfig 描述日志输出端的配置。
//
// 中文说明：
// - Driver 决定输出目标类型：stdout / single / rotate。
// - Filename 是当前日志文件路径；single 模式直接写它，rotate 模式则通常把它作为当前软链接名。
// - Rotate* 与 Lumberjack 参数分别服务于“按时间滚动”和“按大小滚动”两套实现。
type sinkConfig struct {
	Path string

	// Driver can be: stdout|single|rotate
	Driver string

	// Single file (single)
	Filename string

	// Rotate file (rotate)
	RotatePattern string
	RotateMaxAge  time.Duration
	RotateTime    time.Duration

	// Lumberjack options (single)
	MaxSizeMB   int
	MaxBackups  int
	MaxAgeDays  int
	Compress    bool
	LocalTime   bool
}

// buildWriteSyncer 根据 sinkConfig 构造 zap 的输出目标。
//
// driver 说明：
// - stdout：输出到标准输出（开发环境常用）
// - single：输出到单文件，并使用 lumberjack 按“文件大小”滚动
// - rotate：按时间滚动（例如每天一个文件），并保留一个软链接指向当前文件（WithLinkName）
func buildWriteSyncer(sc sinkConfig) (zapcore.WriteSyncer, error) {
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
		// Example pattern: /path/app-%Y%m%d.log
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
	// 中文说明：
	// - 这里传入的是日志文件路径，不是目录路径。
	// - 因此需要先取 filepath.Dir，再递归创建父目录。
	return os.MkdirAll(filepath.Dir(p), 0o755)
}

// CloseIfPossible closes writer if it is io.Closer.
func CloseIfPossible(w any) {
	if c, ok := w.(io.Closer); ok {
		_ = c.Close()
	}
}
