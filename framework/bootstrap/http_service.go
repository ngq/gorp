// Package bootstrap HTTP 服务启动封装
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework"
	"github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
	frameworklog "github.com/ngq/gorp/framework/log"
	gingin "github.com/ngq/gorp/framework/provider/gin"
	"github.com/ngq/gorp/framework/provider/host"
	redisProvider "github.com/ngq/gorp/framework/provider/redis"

	gormpkg "gorm.io/gorm"
)

// HTTPServiceOptions HTTP 服务启动选项。
//
// 中文说明：
// - 这一层只暴露 starter 默认启动真正需要的少量开关；
// - 通过 Disable* 控制是否挂载可选基础能力，避免业务 main.go 再去手工拼 provider 列表；
// - ExtraProviders 用来追加服务私有能力，例如 RPC client、领域仓储或自定义中间件 provider。
type HTTPServiceOptions struct {
	// ExtraProviders 服务专属 Provider 列表。
	ExtraProviders []contract.ServiceProvider

	// DisableRedis 禁用 Redis Provider。
	DisableRedis bool

	// DisableGorm 禁用 Gorm Provider。
	DisableGorm bool

	// DisableMetrics 禁用 Prometheus 指标采集。
	DisableMetrics bool
}

// HTTPServiceRuntime HTTP 服务运行时
//
// 中文说明：
// - 封装 HTTP 服务启动阶段需要的公共运行时对象；
// - 包含框架核心能力：Logger、Engine、Config、DB；
// - 同时暴露 starter 最常用的起步能力：JWT、Redis；
// - 用于简化业务服务的 main.go 启动代码。
type HTTPServiceRuntime struct {
	App         *framework.Application
	Container   contract.Container
	Logger      contract.Logger
	Engine      *gin.Engine
	DB          *gormpkg.DB
	Redis       contract.Redis
	JWT         contract.JWTService
	Config      contract.Config
	ServiceName string
}

// NewHTTPServiceRuntime 创建 HTTP 服务运行时
//
// 中文说明：
// - 初始化框架并返回运行时对象；
// - 自动注册默认 Provider 组；
// - 支持通过选项定制：追加 Provider 或禁用某些能力。
func NewHTTPServiceRuntime(serviceName string, opts HTTPServiceOptions) (*HTTPServiceRuntime, error) {
	app := framework.NewApplication()
	c := app.Container()

	// 构建 Provider 列表
	providers := buildHTTPProviders(opts)
	if err := c.RegisterProviders(providers...); err != nil {
		return nil, fmt.Errorf("注册 Provider 失败: %w", err)
	}
	if err := RegisterSelectedMicroserviceProviders(c); err != nil {
		return nil, fmt.Errorf("注册主链路能力失败: %w", err)
	}

	rt := &HTTPServiceRuntime{
		App:         app,
		Container:   c,
		Logger:      container.MustMakeLogger(c),
		Engine:      container.MustMakeEngine(c),
		Config:      container.MustMakeConfig(c),
		JWT:         container.MustMakeJWTService(c),
		ServiceName: serviceName,
	}
	frameworklog.SetDefault(rt.Logger)

	if !opts.DisableGorm {
		rt.DB = container.MustMakeGorm(c)
	}
	if !opts.DisableRedis {
		if redisSvc, err := container.MakeRedis(c); err == nil {
			rt.Redis = redisSvc
		}
	}

	return rt, nil
}

// buildHTTPProviders 构建 HTTP 服务 Provider 列表。
//
// 中文说明：
// - 先装入 framework 默认基础能力，再按开关决定是否接入 ORM runtime 与 Redis；
// - 业务减负能力保持默认开启，让 starter 生成项目开箱即可拿到 JWT 等常用能力；
// - ExtraProviders 始终最后追加，便于业务在不覆盖默认骨架的前提下补自己的 provider。
func buildHTTPProviders(opts HTTPServiceOptions) []contract.ServiceProvider {
	providers := make([]contract.ServiceProvider, 0)

	// 默认业务起步骨架
	providers = append(providers, FoundationProviders()...)

	// ORM/Runtime
	if !opts.DisableGorm {
		providers = append(providers, ORMRuntimeProviders()...)
	}

	// 默认业务减负能力
	providers = append(providers, DefaultCapabilityProviders()...)
	if !opts.DisableRedis {
		providers = append(providers, redisProvider.NewProvider())
	}

	// 服务专属
	providers = append(providers, opts.ExtraProviders...)

	return providers
}

// BootHTTPService 启动 HTTP 服务。
//
// 中文说明：
// - 这是 generated project 默认公开的 HTTP 启动骨架；
// - 统一处理：初始化 -> 可选迁移 -> 服务装配回调 -> 健康检查 -> Prometheus 指标 -> RunHTTP；
// - 业务项目只需要提供自己的装配逻辑，不需要先理解 runtime provider 或 legacy CLI；
// - 自动注册 /healthz 和 /metrics 端点。
//
// 使用示例：
//
//	err := bootstrap.BootHTTPService("user-service", bootstrap.HTTPServiceOptions{}, migrate, setup)
func BootHTTPService(serviceName string, opts HTTPServiceOptions, migrate func(*HTTPServiceRuntime) error, setup func(*HTTPServiceRuntime) error) error {
	rt, err := NewHTTPServiceRuntime(serviceName, opts)
	if err != nil {
		return fmt.Errorf("初始化失败: %w", err)
	}

	rt.Logger.Info(fmt.Sprintf("%s 正在启动...", serviceName))

	// 执行数据库迁移
	if migrate != nil {
		if err := migrate(rt); err != nil {
			return fmt.Errorf("表结构迁移失败: %w", err)
		}
		rt.Logger.Info("表结构迁移完成")
	}

	// 执行服务装配
	if setup != nil {
		if err := setup(rt); err != nil {
			return fmt.Errorf("服务装配失败: %w", err)
		}
	}

	// 注册基础端点
	RegisterHealthCheck(rt.Engine, serviceName)
	if !opts.DisableMetrics {
		RegisterMetricsEndpoint(rt.Engine)
		rt.Engine.Use(gingin.MetricsMiddleware())
	}

	return RunHTTP(rt.Container, rt.Logger)
}

// AutoMigrateModels 在 HTTP 服务运行时上执行最小 GORM 自动迁移。
//
// 中文说明：
// - 这是 starter/main.go 常见迁移样板的统一收口辅助；
// - rt.DB 为空时直接跳过，便于复用到可选数据库能力路径；
// - 用于减少模板里重复的 `if rt.DB == nil { return nil }` 胶水代码。
func AutoMigrateModels(rt *HTTPServiceRuntime, models ...any) error {
	if rt == nil || rt.DB == nil || len(models) == 0 {
		return nil
	}
	return rt.DB.AutoMigrate(models...)
}

// RegisterHealthCheck 注册健康检查端点
//
// 中文说明：
// - 添加标准 /healthz 端点，返回服务状态；
// - serviceName 用于标识当前服务名称。
func RegisterHealthCheck(engine *gin.Engine, serviceName string) {
	engine.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": serviceName,
			"version": "1.0.0",
		})
	})
}

// RegisterMetricsEndpoint 注册 Prometheus 指标端点
//
// 中文说明：
// - 暴露 /metrics 端点供 Prometheus 采集；
// - 自动收集 HTTP 请求指标；
// - 包含 Go runtime 指标。
func RegisterMetricsEndpoint(engine *gin.Engine) {
	// 注册 Go runtime 指标
	gingin.RegisterGoRuntimeMetrics()
	// 注册 /metrics 端点
	engine.GET("/metrics", gingin.PrometheusHandler())
}

// RunHTTP 启动 HTTP 服务，封装信号处理和优雅关闭
//
// 中文说明：
// - 自动从容器获取 Host 和 HTTP 服务；
// - 注册信号处理（SIGINT、SIGTERM）；
// - 支持优雅关闭，超时时间 10 秒；
// - 如果 Host 未注册，则直接启动 HTTP（不依赖 Host）。
func RunHTTP(c contract.Container, logger contract.Logger) error {
	// 尝试获取 Host 服务
	hostSvc, err := container.MakeHost(c)
	if err != nil {
		// Host 未注册，直接启动 HTTP
		return runHTTPDirectly(c, logger)
	}

	// 获取 HTTP 服务
	httpSvc, err := container.MakeHTTP(c)
	if err != nil {
		return fmt.Errorf("获取 HTTP 服务失败: %w", err)
	}

	// 注册到 Host
	httpHostable := host.NewHTTPService("http", httpSvc)
	if err := hostSvc.RegisterService("http", httpHostable); err != nil {
		return fmt.Errorf("注册 HTTP 服务到 Host 失败: %w", err)
	}

	logger.Info("starting http server")
	if err := hostSvc.Start(context.Background()); err != nil {
		return err
	}

	// 设置信号处理
	sigs := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		sigs = append(sigs, syscall.SIGTERM)
	}
	ctx, stop := signal.NotifyContext(context.Background(), sigs...)
	defer stop()
	<-ctx.Done()

	logger.Info("shutdown signal received")

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := hostSvc.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("优雅关闭失败: %w", err)
	}

	logger.Info("http server stopped gracefully")
	return nil
}

// runHTTPDirectly 直接启动 HTTP 服务（不依赖 Host）。
//
// 中文说明：
// - 这是没有注册 Host 时的最小直启路径；
// - 仍然保留信号监听与优雅关闭，保证 starter 在精简模式下也具备完整退出行为；
// - 该分支只是不经过生命周期编排，不代表框架默认不推荐 Host。
func runHTTPDirectly(c contract.Container, logger contract.Logger) error {
	httpSvc, err := container.MakeHTTP(c)
	if err != nil {
		return fmt.Errorf("获取 HTTP 服务失败: %w", err)
	}

	logger.Info("starting http server (direct mode)")

	// 设置信号处理
	sigs := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		sigs = append(sigs, syscall.SIGTERM)
	}
	ctx, stop := signal.NotifyContext(context.Background(), sigs...)
	defer stop()

	// 启动 HTTP 服务
	errCh := make(chan error, 1)
	go func() {
		if err := httpSvc.Run(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				errCh <- nil
				return
			}
			errCh <- err
		}
	}()

	// 等待信号或错误
	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return err
		}
		return nil
	}

	// 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpSvc.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("优雅关闭失败: %w", err)
	}
	logger.Info("http server stopped gracefully")
	return nil
}
