// Package config 服务配置管理
package config

import (
	"os"
	"strconv"
)

// ServiceConfig 服务基础配置
type ServiceConfig struct {
	// 服务基本信息
	Name    string
	Version string
	Port    int

	// 数据库配置
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	// Redis 配置
	RedisHost string
	RedisPort int

	// 服务发现配置
	ConsulAddr      string
	DiscoveryEnabled bool

	// 链路追踪配置
	TracingEnabled  bool
	TracingEndpoint string
	TracingSampler  float64

	// 监控配置
	MetricsEnabled bool
	MetricsPath    string

	// JWT 配置
	JWTSecret string
}

// LoadFromEnv 从环境变量加载配置
//
// 中文说明：
// - 从环境变量读取服务配置；
// - 提供默认值；
// - 支持容器化部署。
func LoadFromEnv() *ServiceConfig {
	return &ServiceConfig{
		Name:    getEnvOrDefault("SERVICE_NAME", "unknown-service"),
		Version: getEnvOrDefault("SERVICE_VERSION", "1.0.0"),
		Port:    getEnvIntOrDefault("PORT", 8000),

		DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
		DBPort:     getEnvIntOrDefault("DB_PORT", 3306),
		DBUser:     getEnvOrDefault("DB_USER", "root"),
		DBPassword: getEnvOrDefault("DB_PASSWORD", ""),
		DBName:     getEnvOrDefault("DB_NAME", "nop_db"),

		RedisHost: getEnvOrDefault("REDIS_HOST", "localhost"),
		RedisPort: getEnvIntOrDefault("REDIS_PORT", 6379),

		ConsulAddr:      getEnvOrDefault("CONSUL_ADDR", "localhost:8500"),
		DiscoveryEnabled: getEnvBoolOrDefault("DISCOVERY_ENABLED", false),

		TracingEnabled:  getEnvBoolOrDefault("TRACING_ENABLED", false),
		TracingEndpoint: getEnvOrDefault("TRACING_ENDPOINT", "http://jaeger:14268/api/traces"),
		TracingSampler:  getEnvFloatOrDefault("TRACING_SAMPLER", 0.1),

		MetricsEnabled: getEnvBoolOrDefault("METRICS_ENABLED", true),
		MetricsPath:    getEnvOrDefault("METRICS_PATH", "/metrics"),

		JWTSecret: getEnvOrDefault("JWT_SECRET", "your-secret-key"),
	}
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvIntOrDefault 获取整型环境变量或默认值
func getEnvIntOrDefault(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

// getEnvBoolOrDefault 获取布尔型环境变量或默认值
func getEnvBoolOrDefault(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}

// getEnvFloatOrDefault 获取浮点型环境变量或默认值
func getEnvFloatOrDefault(key string, defaultVal float64) float64 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultVal
}