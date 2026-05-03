// Package gorp 提供稳定的 facade 公共入口。
// 说明：业务运行入口始终是项目自己的 main；根包只做薄转发。
package gorp

import (
	"context"

	"github.com/ngq/gorp/framework/contract"
	"github.com/ngq/gorp/framework/facade"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"
)

const DBRuntimeKey = contract.DBRuntimeKey

var (
	// ErrServiceNameRequired 表示未提供服务名。
	ErrServiceNameRequired = facade.ErrServiceNameRequired
	// ErrNoServiceDeclared 表示未声明可运行服务。
	ErrNoServiceDeclared = facade.ErrNoServiceDeclared
	// ErrHTTPRouteRegistrationFailed 表示 HTTP 路由注册失败。
	ErrHTTPRouteRegistrationFailed = facade.ErrHTTPRouteRegistrationFailed
	// ErrHTTPRuntimeUnavailable 表示 HTTP 路由注册阶段缺少可用 runtime。
	ErrHTTPRuntimeUnavailable = facade.ErrHTTPRuntimeUnavailable
	// ErrSetupFailed 表示 setup 回调执行失败。
	ErrSetupFailed = facade.ErrSetupFailed
	// ErrMigrateFailed 表示 migrate 回调执行失败。
	ErrMigrateFailed = facade.ErrMigrateFailed
	// ErrStartupCanceled 表示启动前 context 已取消。
	ErrStartupCanceled = facade.ErrStartupCanceled
	// ErrHTTPServiceRunFailed 表示 HTTP 服务启动失败。
	ErrHTTPServiceRunFailed = facade.ErrHTTPServiceRunFailed
	// ErrHTTPRuntimeBuildFailed 表示 HTTP runtime 构建失败。
	ErrHTTPRuntimeBuildFailed = facade.ErrHTTPRuntimeBuildFailed
)

// HTTPRuntime 是 facade 启动上下文类型别名。
type HTTPRuntime = facade.HTTPRuntime

// HTTPServiceOptions 是默认 HTTP 声明选项类型别名。
type HTTPServiceOptions = facade.HTTPServiceOptions

// Container 是框架容器契约类型别名。
type Container = contract.Container

// HTTPRouter 是默认 HTTP 路由契约类型别名。
type HTTPRouter = contract.HTTPRouter

// HTTPContext 是默认 HTTP handler 上下文类型别名。
type HTTPContext = contract.HTTPContext

// HTTPHandler 是默认 HTTP handler 契约类型别名。
type HTTPHandler = contract.HTTPHandler

// HTTPMiddleware 是默认 HTTP middleware 契约类型别名。
type HTTPMiddleware = contract.HTTPMiddleware

// ServiceProvider 是 provider 声明类型别名。
type ServiceProvider = contract.ServiceProvider

// Metadata 是服务间 metadata 契约类型别名。
type Metadata = contract.Metadata

// GRPCConnFactory 是 Proto-first gRPC 客户端连接工厂类型别名。
type GRPCConnFactory = contract.GRPCConnFactory

// GRPCServerRegistrar 是 Proto-first gRPC 服务端注册器类型别名。
type GRPCServerRegistrar = contract.GRPCServerRegistrar

// DistributedLock 是分布式锁能力契约类型别名。
type DistributedLock = contract.DistributedLock

// MessagePublisher 是消息发布能力契约类型别名。
type MessagePublisher = contract.MessagePublisher

// MessageSubscriber 是消息订阅能力契约类型别名。
type MessageSubscriber = contract.MessageSubscriber

// Message 是消息队列消息对象类型别名。
type Message = contract.Message

// ServiceIdentity 是服务身份信息类型别名。
type ServiceIdentity = contract.ServiceIdentity

// MigrateFunc 是迁移回调契约类型别名。
type MigrateFunc = facade.MigrateFunc

// SetupFunc 是启动装配回调契约类型别名。
type SetupFunc = facade.SetupFunc

// HTTPRouteRegistrar 是默认 HTTP 路由注册契约类型别名。
type HTTPRouteRegistrar = facade.HTTPRouteRegistrar

// Option 是 facade 启动选项类型别名。
type Option = facade.Option

// Run 启动默认 HTTP 主线。
func Run(serviceName string, options ...Option) error {
	return facade.Run(serviceName, options...)
}

// Start 是 Run 的同义入口。
func Start(serviceName string, options ...Option) error {
	return facade.Start(serviceName, options...)
}

// RunContext 使用显式 context 启动默认主线。
func RunContext(ctx context.Context, serviceName string, options ...Option) error {
	return facade.RunContext(ctx, serviceName, options...)
}

// BuildHTTPRuntime 仅构建启动上下文，不启动监听。
func BuildHTTPRuntime(serviceName string, options ...Option) (*HTTPRuntime, error) {
	return facade.BuildHTTPRuntime(serviceName, options...)
}

// Build 是 BuildHTTPRuntime 的同义入口。
func Build(serviceName string, options ...Option) (*HTTPRuntime, error) {
	return facade.Build(serviceName, options...)
}

// HTTP 声明使用默认 HTTP 主线。
func HTTP(opts ...HTTPServiceOptions) Option {
	return facade.HTTP(opts...)
}

// WithoutHTTP 显式关闭默认 HTTP 声明。
func WithoutHTTP() Option {
	return facade.WithoutHTTP()
}

// Module 声明单个模块的 providers。
func Module(providers ...ServiceProvider) Option {
	return facade.Module(providers...)
}

// Modules 声明一组模块 providers。
func Modules(groups ...[]ServiceProvider) Option {
	return facade.Modules(groups...)
}

// WithModule 是 Module 的显式命名入口。
func WithModule(providers ...ServiceProvider) Option {
	return facade.WithModule(providers...)
}

// WithProviders 追加 providers 声明，不改变底层选型语义。
func WithProviders(providers ...ServiceProvider) Option {
	return facade.WithProviders(providers...)
}

// WithMigrate 声明迁移回调。
func WithMigrate(fn MigrateFunc) Option {
	return facade.WithMigrate(fn)
}

// WithSetup 声明装配回调。
func WithSetup(fn SetupFunc) Option {
	return facade.WithSetup(fn)
}

// WithHTTPRoutes 声明默认 HTTP 路由注册回调。
func WithHTTPRoutes(register HTTPRouteRegistrar) Option {
	return facade.WithHTTPRoutes(register)
}

// MakeGRPCConnFactory 获取 Proto-first gRPC 连接工厂。
func MakeGRPCConnFactory(c Container) (GRPCConnFactory, error) {
	return facade.MakeGRPCConnFactory(c)
}

// MakeGRPCServerRegistrar 获取 Proto-first gRPC 服务端注册器。
func MakeGRPCServerRegistrar(c Container) (GRPCServerRegistrar, error) {
	return facade.MakeGRPCServerRegistrar(c)
}

// MakeDistributedLock 获取分布式锁能力。
func MakeDistributedLock(c Container) (DistributedLock, error) {
	return facade.MakeDistributedLock(c)
}

// MakeMessagePublisher 获取消息发布能力。
func MakeMessagePublisher(c Container) (MessagePublisher, error) {
	return facade.MakeMessagePublisher(c)
}

// MakeMessageSubscriber 获取消息订阅能力。
func MakeMessageSubscriber(c Container) (MessageSubscriber, error) {
	return facade.MakeMessageSubscriber(c)
}

// NewMetadata 创建空 metadata。
func NewMetadata() Metadata {
	return contract.NewMetadata()
}

// NewServerContext 创建带服务端 metadata 的上下文。
func NewServerContext(ctx context.Context, md Metadata) context.Context {
	return contract.NewServerContext(ctx, md)
}

// FromServerContext 读取服务端 metadata。
func FromServerContext(ctx context.Context) (Metadata, bool) {
	return contract.FromServerContext(ctx)
}

// NewClientContext 创建带客户端 metadata 的上下文。
func NewClientContext(ctx context.Context, md Metadata) context.Context {
	return contract.NewClientContext(ctx, md)
}

// FromClientContext 读取客户端 metadata。
func FromClientContext(ctx context.Context) (Metadata, bool) {
	return contract.FromClientContext(ctx)
}

// AppendToClientContext 向客户端上下文追加 metadata。
func AppendToClientContext(ctx context.Context, kv ...string) context.Context {
	return contract.AppendToClientContext(ctx, kv...)
}

// WithServiceIdentity 把服务身份写入上下文。
func WithServiceIdentity(ctx context.Context, identity *ServiceIdentity) context.Context {
	return facade.WithServiceIdentity(ctx, identity)
}

// FromServiceIdentity 读取上下文中的服务身份。
func FromServiceIdentity(ctx context.Context) (*ServiceIdentity, bool) {
	return facade.FromServiceIdentity(ctx)
}

// GetGRPCTraceID 从 gRPC context 读取 trace id。
func GetGRPCTraceID(ctx context.Context) string {
	return facade.GetGRPCTraceID(ctx)
}

// GetGRPCRequestID 从 gRPC context 读取 request id。
func GetGRPCRequestID(ctx context.Context) string {
	return facade.GetGRPCRequestID(ctx)
}

func NewContainerContext(ctx context.Context, c Container) context.Context {
	return contract.NewContainerContext(ctx, c)
}

func FromContainerContext(ctx context.Context) (Container, bool) {
	return contract.FromContainerContext(ctx)
}

func FromJWTClaimsContext(ctx context.Context) (*contract.JWTClaims, bool) {
	return contract.FromJWTClaimsContext(ctx)
}

func FromSubjectIDContext(ctx context.Context) (int64, bool) {
	return contract.FromSubjectIDContext(ctx)
}

func FromSubjectTypeContext(ctx context.Context) (string, bool) {
	return contract.FromSubjectTypeContext(ctx)
}

func FromValidatedBodyContext(ctx context.Context) (any, bool) {
	return contract.FromValidatedBodyContext(ctx)
}

func FromRequestIDContext(ctx context.Context) (string, bool) {
	return contract.FromRequestIDContext(ctx)
}

func FromTraceIDContext(ctx context.Context) (string, bool) {
	return contract.FromTraceIDContext(ctx)
}

func Success(c HTTPContext, data any) {
	ginprovider.Success(c, data)
}

func SuccessWithMessage(c HTTPContext, message string, data any) {
	ginprovider.SuccessWithMessage(c, message, data)
}

func SuccessWithStatus(c HTTPContext, status int, data any) {
	ginprovider.SuccessWithStatus(c, status, data)
}

func Error(c HTTPContext, err error) {
	ginprovider.Error(c, err)
}

func BadRequest(c HTTPContext, message string) {
	ginprovider.BadRequest(c, message)
}

func InternalError(c HTTPContext, message string) {
	ginprovider.InternalError(c, message)
}
