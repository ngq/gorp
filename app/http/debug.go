package http

import (
	"context"
	"net/http"
	"net/http/pprof"
	"runtime"

	"github.com/gin-gonic/gin"
	frameworkcontainer "github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"
)

// RegisterPprof 注册 pprof 性能分析端点。
//
// 中文说明：
// - 暴露 Go 标准库 pprof 端点，用于性能分析和问题排查；
// - 包括 CPU profiling、heap profiling、goroutine 分析等；
// - 生产环境建议通过配置控制是否启用，或添加认证保护；
// - 端点路径：
//   - GET /debug/pprof/          - pprof 索引页
//   - GET /debug/pprof/cmdline   - 命令行参数
//   - GET /debug/pprof/profile   - CPU profiling
//   - GET /debug/pprof/symbol    - 符号表
//   - GET /debug/pprof/trace     - 执行追踪
//   - GET /debug/pprof/heap      - 堆内存分析
//   - GET /debug/pprof/goroutine - goroutine 分析
//   - GET /debug/pprof/block     - 阻塞分析
//   - GET /debug/pprof/mutex     - 互斥锁分析
//   - GET /debug/pprof/threadcreate - 线程创建分析
func RegisterPprof(c contract.Container) error {
	engine, err := ginprovider.EngineFromContainer(c)
	if err != nil {
		return err
	}

	// pprof 索引页
	engine.GET("/debug/pprof/", gin.WrapF(pprof.Index))
	// 命令行参数
	engine.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	// CPU profiling
	engine.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	// 符号表
	engine.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	// 执行追踪
	engine.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
	// 堆内存分析
	engine.GET("/debug/pprof/heap", gin.WrapH(pprof.Handler("heap")))
	// goroutine 分析
	engine.GET("/debug/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
	// 阻塞分析（需要先调用 runtime.SetBlockProfileRate）
	engine.GET("/debug/pprof/block", gin.WrapH(pprof.Handler("block")))
	// 互斥锁分析（需要先调用 runtime.SetMutexProfileFraction）
	engine.GET("/debug/pprof/mutex", gin.WrapH(pprof.Handler("mutex")))
	// 线程创建分析
	engine.GET("/debug/pprof/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))

	return nil
}

// RegisterHealthProbes 注册健康探针端点。
//
// 中文说明：
// - /healthz - 存活探针（Liveness Probe），检查服务是否存活；
// - /readyz - 就绪探针（Readiness Probe），检查服务是否准备好接收请求；
// - 用于 Kubernetes 探针和负载均衡健康检查；
// - 就绪探针会检查关键依赖（DB、Redis）是否可用。
func RegisterHealthProbes(c contract.Container) error {
	engine, err := ginprovider.EngineFromContainer(c)
	if err != nil {
		return err
	}

	// 存活探针：检查服务是否存活
	// 中文说明：
	// - 只要服务能响应，就认为存活；
	// - 用于 Kubernetes livenessProbe；
	// - 如果此端点失败，Kubernetes 会重启容器。
	engine.GET("/healthz", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 就绪探针：检查服务是否准备好接收请求
	// 中文说明：
	// - 检查关键依赖（DB、Redis）是否可用；
	// - 用于 Kubernetes readinessProbe；
	// - 如果此端点失败，Kubernetes 会将 Pod 从 Service 端点中移除。
	engine.GET("/readyz", func(ctx *gin.Context) {
		checks := make(map[string]string)
		allHealthy := true

		// 检查数据库连接
		if err := checkDatabase(c); err != nil {
			checks["database"] = "unhealthy: " + err.Error()
			allHealthy = false
		} else {
			checks["database"] = "ok"
		}

		// 检查 Redis 连接（可选）
		if err := checkRedis(c); err != nil {
			checks["redis"] = "unhealthy: " + err.Error()
			// Redis 不是必须的，不影响整体健康状态
		} else {
			checks["redis"] = "ok"
		}

		status := "ok"
		statusCode := http.StatusOK
		if !allHealthy {
			status = "not ready"
			statusCode = http.StatusServiceUnavailable
		}

		ctx.JSON(statusCode, gin.H{
			"status": status,
			"checks": checks,
		})
	})

	return nil
}

// checkDatabase 检查数据库连接是否正常。
func checkDatabase(c contract.Container) error {
	// 尝试获取统一数据库运行时并执行简单 ping。
	dbAny, err := frameworkcontainer.MakeDBRuntime(c)
	if err != nil {
		return err
	}
	return frameworkcontainer.PingDBRuntime(dbAny)
}

// checkRedis 检查 Redis 连接是否正常。
func checkRedis(c contract.Container) error {
	redis, err := frameworkcontainer.MakeRedis(c)
	if err != nil {
		return err
	}

	// 中文说明：
	// - readiness 检查必须传入有效 context，不能把 nil 传给底层客户端；
	// - 这里使用短生命周期的 background context，避免探针路径因为 nil context 直接报错；
	// - 后续如果需要更细粒度控制，可再加超时包装。
	return redis.Ping(context.Background())
}

// SetupBlockAndMutexProfiles 启用 block 和 mutex profile 采样。
//
// 中文说明：
// - 默认情况下 block 和 mutex profile 是关闭的；
// - 调用此函数启用采样，用于分析阻塞和锁竞争问题；
// - rate 参数：
//   - block rate: 采样阻塞时间超过该值的操作（纳秒），0 表示关闭
//   - mutex fraction: 采样 mutex 竞争的比例，0 表示关闭
func SetupBlockAndMutexProfiles(blockRate int, mutexFraction int) {
	runtime.SetBlockProfileRate(blockRate)
	runtime.SetMutexProfileFraction(mutexFraction)
}