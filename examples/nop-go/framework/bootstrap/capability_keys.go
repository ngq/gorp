// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file defines governance capability keys and provider backend keys as typed constants,
// replacing scattered hardcoded strings with a single source of truth.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件将治理能力名称和 Provider 后端名称定义为类型化常量，
// 用单一真实来源替代散布在代码中的硬编码字符串。
package bootstrap

// CapabilityKey is a typed string representing a governance capability name.
// Use CapabilityXxx constants instead of raw string literals.
//
// CapabilityKey 是表示治理能力名称的类型化字符串。
// 请使用 CapabilityXxx 常量替代裸字符串字面量。
type CapabilityKey string

// BackendKey is a typed string representing a provider backend name.
// Use BackendXxx constants instead of raw string literals.
//
// BackendKey 是表示 Provider 后端名称的类型化字符串。
// 请使用 BackendXxx 常量替代裸字符串字面量。
type BackendKey string

// ──────────────────────────────────────────────────────
// Governance capability keys
// 治理能力名称常量
// ──────────────────────────────────────────────────────
const (
	// CapabilityConfigSource is the governance key for configuration source.
	// CapabilityConfigSource 是配置源的治理能力名称。
	CapabilityConfigSource CapabilityKey = "configsource"

	// CapabilityDiscovery is the governance key for service discovery.
	// CapabilityDiscovery 是服务发现的治理能力名称。
	CapabilityDiscovery CapabilityKey = "discovery"

	// CapabilitySelector is the governance key for load balancing selector.
	// CapabilitySelector 是负载均衡选择器的治理能力名称。
	CapabilitySelector CapabilityKey = "selector"

	// CapabilityRPC is the governance key for RPC transport.
	// CapabilityRPC 是 RPC 传输的治理能力名称。
	CapabilityRPC CapabilityKey = "rpc"

	// CapabilityTracing is the governance key for distributed tracing.
	// CapabilityTracing 是分布式追踪的治理能力名称。
	CapabilityTracing CapabilityKey = "tracing"

	// CapabilityMetadata is the governance key for metadata propagation.
	// CapabilityMetadata 是元数据透传的治理能力名称。
	CapabilityMetadata CapabilityKey = "metadata"

	// CapabilityServiceAuth is the governance key for service authentication.
	// CapabilityServiceAuth 是服务鉴权的治理能力名称。
	CapabilityServiceAuth CapabilityKey = "serviceauth"

	// CapabilityCircuitBreaker is the governance key for circuit breaking.
	// CapabilityCircuitBreaker 是熔断的治理能力名称。
	CapabilityCircuitBreaker CapabilityKey = "circuitbreaker"

	// CapabilityLoadShedding is the governance key for load shedding.
	// CapabilityLoadShedding 是过载保护的治理能力名称。
	CapabilityLoadShedding CapabilityKey = "loadshedding"

	// CapabilityRetry is the governance key for retry.
	// CapabilityRetry 是重试的治理能力名称。
	CapabilityRetry CapabilityKey = "retry"

	// CapabilityDTM is the governance key for distributed transaction manager.
	// CapabilityDTM 是分布式事务管理器的治理能力名称。
	CapabilityDTM CapabilityKey = "dtm"

	// CapabilityMessageQueue is the governance key for message queue.
	// CapabilityMessageQueue 是消息队列的治理能力名称。
	CapabilityMessageQueue CapabilityKey = "message_queue"

	// CapabilityDistributedLock is the governance key for distributed lock.
	// CapabilityDistributedLock 是分布式锁的治理能力名称。
	CapabilityDistributedLock CapabilityKey = "distributed_lock"

	// CapabilityWebSocket is the governance key for WebSocket.
	// CapabilityWebSocket 是 WebSocket 的治理能力名称。
	CapabilityWebSocket CapabilityKey = "websocket"
)

// AllCapabilities returns all defined governance capability keys.
// Use this to iterate over capabilities without repeating the list.
//
// AllCapabilities 返回所有已定义的治理能力名称常量。
// 用于遍历所有能力而无需重复列表。
func AllCapabilities() []CapabilityKey {
	return []CapabilityKey{
		CapabilityConfigSource,
		CapabilityDiscovery,
		CapabilitySelector,
		CapabilityRPC,
		CapabilityTracing,
		CapabilityMetadata,
		CapabilityServiceAuth,
		CapabilityCircuitBreaker,
		CapabilityLoadShedding,
		CapabilityRetry,
		CapabilityDTM,
		CapabilityMessageQueue,
		CapabilityDistributedLock,
		CapabilityWebSocket,
	}
}

// String implements fmt.Stringer.
func (k CapabilityKey) String() string { return string(k) }

// String implements fmt.Stringer.
func (k BackendKey) String() string { return string(k) }

// ──────────────────────────────────────────────────────
// Provider backend keys
// Provider 后端名称常量
// ──────────────────────────────────────────────────────
const (
	// ── Common backends ────────────────────────────────
	// BackendNoop is the no-op backend shared by all capabilities.
	// BackendNoop 是所有能力共享的空实现后端。
	BackendNoop BackendKey = "noop"

	// BackendDefault is the default implementation backend.
	// BackendDefault 是默认实现后端。
	BackendDefault BackendKey = "default"

	// BackendLocal is the local implementation backend.
	// BackendLocal 是本地实现后端。
	BackendLocal BackendKey = "local"

	// ── ConfigSource backends ──────────────────────────
	BackendConsul  BackendKey = "consul"
	BackendEtcd    BackendKey = "etcd"
	BackendApollo  BackendKey = "apollo"
	BackendNacos   BackendKey = "nacos"
	BackendK8s     BackendKey = "kubernetes"
	BackendPolaris BackendKey = "polaris"

	// ── Discovery backends (shared with ConfigSource) ──
	BackendZookeeper   BackendKey = "zookeeper"
	BackendEureka      BackendKey = "eureka"
	BackendServiceComb BackendKey = "servicecomb"

	// ── Selector backends ──────────────────────────────
	BackendRandom  BackendKey = "random"
	BackendWRR     BackendKey = "wrr"
	BackendP2C     BackendKey = "p2c"
	BackendP2CEWMA BackendKey = "p2c_ewma"

	// ── RPC backends ───────────────────────────────────
	BackendHTTP BackendKey = "http"
	BackendGRPC BackendKey = "grpc"

	// ── Tracing backends ───────────────────────────────
	BackendOtel   BackendKey = "otel"
	BackendOTLP   BackendKey = "otlp"
	BackendStdout BackendKey = "stdout"

	// ── ServiceAuth backends ────────────────────────────
	BackendToken BackendKey = "token"
	BackendMTLS  BackendKey = "mtls"

	// ── CircuitBreaker backends ─────────────────────────
	BackendSentinel BackendKey = "sentinel"

	// ── LoadShedding backends ───────────────────────────
	BackendSemaphore BackendKey = "semaphore"

	// ── DTM backends ────────────────────────────────────
	BackendSDK    BackendKey = "sdk"
	BackendDTMSDK BackendKey = "dtmsdk"

	// ── MessageQueue backends ───────────────────────────
	BackendRedis    BackendKey = "redis"
	BackendKafka    BackendKey = "kafka"
	BackendRabbitMQ BackendKey = "rabbitmq"
	BackendRocketMQ BackendKey = "rocketmq"

	// ── WebSocket backends ─────────────────────────────
	BackendGWS BackendKey = "gws"
)

// ──────────────────────────────────────────────────────
// Feature keys (used in governance.enable / governance.disable)
// 特性名称常量（用于 governance.enable / governance.disable）
// ──────────────────────────────────────────────────────
const (
	FeatureRequestIdentity = "request_identity"
	FeatureLogging         = "logging"
	FeatureRecovery        = "recovery"
	FeatureTimeout         = "timeout"
	FeatureMetrics         = "metrics"
	FeatureMetadata        = "metadata"
	FeatureTracing         = "tracing"
	FeatureSelector        = "selector"
	FeatureServiceAuth     = "serviceauth"
	FeatureCircuitBreaker  = "circuitbreaker"
	FeatureRetry           = "retry"
	FeatureLoadShedding    = "loadshedding"
	FeatureDiscovery       = "discovery"
)
