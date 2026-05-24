// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file assembles and runs the gRPC service alongside the HTTP mainline.
// Builds a reusable runtime carrying container, gRPC server, config, and logger.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件装配并运行与 HTTP 主线并行的 gRPC 服务。
// 构建复用型 runtime 对象，统一承载 container、gRPC server、config、logger 能力。
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ngq/gorp/framework"
	"github.com/ngq/gorp/framework/container"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	observabilitycontract "github.com/ngq/gorp/framework/contract/observability"
	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
	frameworklog "github.com/ngq/gorp/framework/log"
	"google.golang.org/grpc"
)

// GRPCServiceOptions describes the bootstrap options for the gRPC service mainline.
//
// GRPCServiceOptions 描述 gRPC 服务主线的 bootstrap 选项。
type GRPCServiceOptions struct {
	// ExtraProviders 是额外追加到容器中的 provider 列表。
	//
	// ExtraProviders 是额外追加到容器中的 provider 列表。
	ExtraProviders []runtimecontract.ServiceProvider

	// GovernanceMode 显式声明运行时治理模式。
	//
	// GovernanceMode 显式声明运行时治理模式。
	GovernanceMode string

	// GovernanceDisable 显式关闭的默认治理能力列表。
	//
	// GovernanceDisable 显式关闭的默认治理能力列表。
	GovernanceDisable []string

	// GovernanceEnable 显式开启的默认关闭治理能力列表。
	//
	// GovernanceEnable 显式开启的默认关闭治理能力列表。
	GovernanceEnable []string

	// GovernanceProviders 显式覆盖的治理 provider backend 映射。
	//
	// GovernanceProviders 显式覆盖的治理 provider backend 映射。
	GovernanceProviders map[string]string
}

// GRPCServiceRuntime carries the assembled gRPC runtime state used during startup callbacks.
//
// GRPCServiceRuntime 承载启动回调阶段使用的 gRPC runtime 状态。
type GRPCServiceRuntime struct {
	// App 是框架 Application 实例。
	//
	// App 是框架 Application 实例。
	App *framework.Application

	// Container 是运行时依赖注入容器。
	//
	// Container 是运行时依赖注入容器。
	Container runtimecontract.Container

	// Logger 是日志能力实例。
	//
	// Logger 是日志能力实例。
	Logger observabilitycontract.Logger

	// Server 是底层 gRPC 服务器实例，供用户在 setup 阶段注册 proto 服务。
	//
	// Server 是底层 gRPC 服务器实例，供用户在 setup 阶段注册 proto 服务。
	Server *grpc.Server

	// Registrar 是 gRPC 服务注册器，提供 RegisterProto 和 Server 方法。
	//
	// Registrar 是 gRPC 服务注册器，提供 RegisterProto 和 Server 方法。
	Registrar transportcontract.GRPCServerRegistrar

	// Config 是配置能力实例。
	//
	// Config 是配置能力实例。
	Config datacontract.Config

	// ServiceName 是当前服务名。
	//
	// ServiceName 是当前服务名。
	ServiceName string

	// GovernanceMode 是生效的治理模式。
	//
	// GovernanceMode 是生效的治理模式。
	GovernanceMode resiliencecontract.GovernanceMode

	// GovernanceSummary 是治理摘要信息。
	//
	// GovernanceSummary 是治理摘要信息。
	GovernanceSummary GovernanceSummary
}

// NewGRPCServiceRuntime builds the default gRPC runtime without starting the server.
//
// NewGRPCServiceRuntime 构建默认 gRPC runtime，但不启动服务。
func NewGRPCServiceRuntime(serviceName string, opts GRPCServiceOptions) (*GRPCServiceRuntime, error) {
	app := framework.NewApplication()
	c := app.Container()

	// 组装基础 provider 列表（与 HTTP 主线共享 foundation + capability）
	// Assemble base provider list (shared foundation + capabilities with HTTP mainline)
	providers := buildGRPCProviders(opts)
	if err := c.RegisterProviders(providers...); err != nil {
		return nil, fmt.Errorf("register providers: %w", err)
	}

	// 注册治理模式选择出的微服务能力 provider
	// Register microservice capability providers selected by governance mode
	if err := registerSelectedMicroserviceProvidersWithOptionsFunc(c, opts.GovernanceMode, opts.GovernanceDisable, opts.GovernanceEnable, opts.GovernanceProviders); err != nil {
		return nil, fmt.Errorf("register selected microservice providers: %w", err)
	}

	// 从容器获取 gRPC 服务注册器
	// Resolve gRPC server registrar from container
	registrar, err := container.MakeGRPCServerRegistrar(c)
	if err != nil {
		return nil, fmt.Errorf("make grpc server registrar: %w", err)
	}

	rt := &GRPCServiceRuntime{
		App:         app,
		Container:   c,
		Logger:      container.MustMakeLogger(c),
		Registrar:   registrar,
		Server:      registrar.Server(),
		Config:      container.MustMakeConfig(c),
		ServiceName: serviceName,
	}
	frameworklog.SetDefault(rt.Logger)
	container.SetDefault(rt.Container)

	// 解析治理模式与摘要
	// Resolve governance mode and summary
	effectiveConfig := overlayGovernanceConfig(rt.Config, opts.GovernanceDisable, opts.GovernanceEnable, opts.GovernanceProviders)
	governanceMode := DetectGovernanceMode(effectiveConfig)
	if opts.GovernanceMode != "" {
		governanceMode = NormalizeGovernanceMode(resiliencecontract.GovernanceMode(opts.GovernanceMode))
	}
	governanceSummary := BuildGovernanceSummaryWithModeOverride(effectiveConfig, governanceMode, opts.GovernanceMode)
	rt.GovernanceMode = governanceMode
	rt.GovernanceSummary = governanceSummary
	rt.Logger.Info(FormatGovernanceSummary(governanceSummary))

	return rt, nil
}

// buildGRPCProviders assembles the provider list used by the gRPC service mainline.
//
// buildGRPCProviders 组装 gRPC 服务主线使用的 provider 列表。
func buildGRPCProviders(opts GRPCServiceOptions) []runtimecontract.ServiceProvider {
	providers := make([]runtimecontract.ServiceProvider, 0)
	providers = append(providers, FoundationProviders()...)
	providers = append(providers, ORMRuntimeProviders()...)
	providers = append(providers, DefaultCapabilityProviders()...)
	providers = append(providers, opts.ExtraProviders...)
	return providers
}

// BootGRPCService assembles, configures, and runs the standalone gRPC service.
// This is the gRPC counterpart of BootHTTPService, used when gRPC runs independently.
//
// BootGRPCService 装配、配置并运行独立的 gRPC 服务。
// 这是 BootHTTPService 的 gRPC 对称入口，在 gRPC 独立运行时使用。
func BootGRPCService(serviceName string, opts GRPCServiceOptions, setup func(*GRPCServiceRuntime) error) error {
	rt, err := NewGRPCServiceRuntime(serviceName, opts)
	if err != nil {
		return fmt.Errorf("initialize grpc runtime: %w", err)
	}

	rt.Logger.Info(fmt.Sprintf("%s starting (gRPC)", serviceName))

	// 执行 setup 回调，让用户注册 gRPC 服务
	// Execute setup callback for user to register gRPC services
	if setup != nil {
		if err := setup(rt); err != nil {
			return fmt.Errorf("setup service: %w", err)
		}
	}

	return RunGRPC(rt.Container, rt.Logger)
}

// RunGRPC runs the gRPC service through the host capability when present.
// Falls back to direct gRPC run mode when host is unavailable.
//
// RunGRPC 优先通过 host 能力运行 gRPC 服务。
// 当 host 能力不可用时，回退到 gRPC 直跑模式。
func RunGRPC(c runtimecontract.Container, logger observabilitycontract.Logger) error {
	// 从容器解析 RPC Server（gRPC 实现）
	// Resolve RPC Server (gRPC implementation) from container
	rpcServerAny, err := c.Make(transportcontract.RPCServerKey)
	if err != nil {
		return fmt.Errorf("make rpc server: %w", err)
	}
	rpcServer, ok := rpcServerAny.(transportcontract.RPCServer)
	if !ok {
		return errors.New("rpc server does not implement transportcontract.RPCServer")
	}

	// 优先尝试通过 host 能力运行
	// Try running through host capability first
	hostSvc, err := container.MakeHost(c)
	if err == nil {
		// 通过 host 管理生命周期
		// Manage lifecycle through host
		grpcHostable, err := newGRPCHostableFromRPCServer(rpcServer)
		if err != nil {
			return fmt.Errorf("create grpc hostable: %w", err)
		}
		if err := hostSvc.RegisterService("grpc", grpcHostable); err != nil {
			return fmt.Errorf("register grpc service to host: %w", err)
		}

		logger.Info("starting grpc server")
		if err := hostSvc.Start(context.Background()); err != nil {
			return err
		}

		sigs := []os.Signal{os.Interrupt}
		if runtime.GOOS != "windows" {
			sigs = append(sigs, syscall.SIGTERM)
		}
		ctx, stop := signal.NotifyContext(context.Background(), sigs...)
		defer stop()
		<-ctx.Done()

		logger.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := hostSvc.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("graceful shutdown failed: %w", err)
		}

		logger.Info("grpc server stopped gracefully")
		return nil
	}

	// 回退到 gRPC 直跑模式
	// Fall back to direct gRPC run mode
	return runGRPCDirectly(rpcServer, logger)
}

// runGRPCDirectly runs the gRPC server without the host abstraction.
//
// runGRPCDirectly 在不使用 host 抽象的情况下直接运行 gRPC 服务。
func runGRPCDirectly(rpcServer transportcontract.RPCServer, logger observabilitycontract.Logger) error {
	logger.Info("starting grpc server (direct mode)")

	sigs := []os.Signal{os.Interrupt}
	if runtime.GOOS != "windows" {
		sigs = append(sigs, syscall.SIGTERM)
	}
	ctx, stop := signal.NotifyContext(context.Background(), sigs...)
	defer stop()

	if err := rpcServer.Start(ctx); err != nil {
		return fmt.Errorf("grpc server start failed: %w", err)
	}

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := rpcServer.Stop(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("grpc server stopped gracefully")
	return nil
}

// newGRPCHostableFromRPCServer creates a Hostable adapter from an RPCServer.
// This bridges the RPCServer contract with the Host's Hostable interface.
//
// newGRPCHostableFromRPCServer 从 RPCServer 创建 Hostable 适配器。
// 将 RPCServer 契约桥接到 Host 的 Hostable 接口。
func newGRPCHostableFromRPCServer(rpcServer transportcontract.RPCServer) (runtimecontract.Hostable, error) {
	// 如果 RPCServer 实现了 GRPCServerRegistrar 接口，可以直接获取底层 grpc.Server
	// If RPCServer also implements GRPCServerRegistrar, we can get the underlying grpc.Server
	type grpcServerRegistrar interface {
		Server() *grpc.Server
		GRPCServer() *grpc.Server
	}

	registrar, ok := rpcServer.(grpcServerRegistrar)
	if !ok {
		return nil, errors.New("rpc server does not expose gRPC server instance")
	}

	// 启动 gRPC 服务获取监听地址，然后通过 host 管理
	// Start the gRPC server to get listener, then manage through host
	grpcServer := registrar.GRPCServer()
	return &rpcServerHostable{
		name:   "grpc",
		server: rpcServer,
		grpc:   grpcServer,
	}, nil
}

// rpcServerHostable adapts transportcontract.RPCServer to runtimecontract.Hostable.
//
// rpcServerHostable 将 transportcontract.RPCServer 适配为 runtimecontract.Hostable。
type rpcServerHostable struct {
	name   string
	server transportcontract.RPCServer
	grpc   *grpc.Server
}

// Name 返回服务名称。
func (h *rpcServerHostable) Name() string { return h.name }

// Start 启动 gRPC 服务器。
func (h *rpcServerHostable) Start(ctx context.Context) error {
	return h.server.Start(ctx)
}

// Stop 优雅停止 gRPC 服务器。
func (h *rpcServerHostable) Stop(ctx context.Context) error {
	return h.server.Stop(ctx)
}

// StartGRPCServer starts the gRPC server from the container when the container
// has a GRPCServerRegistrar bound. Returns the RPCServer if started, nil otherwise.
// This is used by the HTTP bootstrap to auto-start gRPC alongside HTTP in microservice mode.
//
// StartGRPCServer 在容器中存在 GRPCServerRegistrar 时启动 gRPC 服务器。
// 返回启动的 RPCServer；如果容器中没有 gRPC 能力则返回 nil。
// 用于 HTTP bootstrap 在微服务模式下自动与 HTTP 一起启动 gRPC。
func StartGRPCServer(c runtimecontract.Container, logger observabilitycontract.Logger) (transportcontract.RPCServer, error) {
	// 检查容器是否绑定了 gRPC 服务注册器
	// Check if container has gRPC server registrar bound
	if !c.IsBind(transportcontract.GRPCServerRegistrarKey) {
		return nil, nil
	}

	// 解析 RPC Server
	// Resolve RPC Server
	rpcServerAny, err := c.Make(transportcontract.RPCServerKey)
	if err != nil {
		// 容器中有 GRPCServerRegistrar 但解析 RPCServer 失败，说明可能使用了 noop provider
		// Container has GRPCServerRegistrar but RPCServer resolution failed, likely using noop provider
		logger.Info("grpc server not available, skipping grpc startup")
		return nil, nil
	}

	rpcServer, ok := rpcServerAny.(transportcontract.RPCServer)
	if !ok {
		logger.Info("rpc server is not a gRPC server, skipping grpc startup")
		return nil, nil
	}

	// 检查是否为 noop 实现（noop server 的 Start 会返回错误）
	// Check if it's a noop implementation (noop server's Start returns error)
	if rpcServer.Addr() == "" && !c.IsBind(transportcontract.GRPCConnFactoryKey) {
		// 可能是 noop provider，不启动
		// Likely noop provider, don't start
		return nil, nil
	}

	return rpcServer, nil
}
