package bootstrap

import (
	"runtime"
	"strings"
	"testing"

	resiliencecontract "github.com/ngq/gorp/framework/contract/resilience"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildGovernanceDefaultsTableMonolithFeatureDefaults 验证单体模式的 feature 默认值。
func TestBuildGovernanceDefaultsTableMonolithFeatureDefaults(t *testing.T) {
	table := BuildGovernanceDefaultsTable(resiliencecontract.GovernanceModeMonolith)
	require.NotNil(t, table)
	assert.Equal(t, resiliencecontract.GovernanceModeMonolith, table.Mode)

	// 单体模式：5 个基础能力默认启用
	assert.True(t, table.FeatureDefaults["request_identity"])
	assert.True(t, table.FeatureDefaults["logging"])
	assert.True(t, table.FeatureDefaults["recovery"])
	assert.True(t, table.FeatureDefaults["timeout"])
	assert.True(t, table.FeatureDefaults["metrics"])

	// 单体模式：微服务增强能力默认关闭
	assert.False(t, table.FeatureDefaults["metadata"])
	assert.False(t, table.FeatureDefaults["tracing"])
	assert.False(t, table.FeatureDefaults["selector"])
	assert.False(t, table.FeatureDefaults["serviceauth"])
	assert.False(t, table.FeatureDefaults["circuitbreaker"])

	// 尚未进入默认主线的治理能力
	assert.False(t, table.FeatureDefaults["retry"])
	assert.False(t, table.FeatureDefaults["loadshedding"])
	assert.False(t, table.FeatureDefaults["discovery"])
}

// TestBuildGovernanceDefaultsTableMicroserviceFeatureDefaults 验证微服务模式的 feature 默认值。
func TestBuildGovernanceDefaultsTableMicroserviceFeatureDefaults(t *testing.T) {
	table := BuildGovernanceDefaultsTable(resiliencecontract.GovernanceModeMicroservice)
	require.NotNil(t, table)
	assert.Equal(t, resiliencecontract.GovernanceModeMicroservice, table.Mode)

	// 微服务模式：5 个基础能力 + 6 个增强能力默认启用（包括 loadshedding）
	assert.True(t, table.FeatureDefaults["request_identity"])
	assert.True(t, table.FeatureDefaults["logging"])
	assert.True(t, table.FeatureDefaults["recovery"])
	assert.True(t, table.FeatureDefaults["timeout"])
	assert.True(t, table.FeatureDefaults["metrics"])
	assert.True(t, table.FeatureDefaults["metadata"])
	assert.True(t, table.FeatureDefaults["tracing"])
	assert.True(t, table.FeatureDefaults["selector"])
	assert.True(t, table.FeatureDefaults["serviceauth"])
	assert.True(t, table.FeatureDefaults["circuitbreaker"])
	assert.True(t, table.FeatureDefaults["loadshedding"])

	// 尚未进入默认主线的治理能力
	assert.False(t, table.FeatureDefaults["retry"])
	assert.False(t, table.FeatureDefaults["discovery"])
}

// TestBuildGovernanceDefaultsTableProviderDefaultsConsistency 验证三种模式 provider 默认值与代码一致。
func TestBuildGovernanceDefaultsTableProviderDefaultsConsistency(t *testing.T) {
	modes := []resiliencecontract.GovernanceMode{
		resiliencecontract.GovernanceModeMonolith,
		resiliencecontract.GovernanceModeMicroservice,
		resiliencecontract.GovernanceModeGinFirst,
	}
	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			table := BuildGovernanceDefaultsTable(mode)
			require.NotNil(t, table)

			// 与 DefaultGovernanceProviderDefaults 返回值一致
			expected := governanceProviderMap(DefaultGovernanceProviderDefaults(mode))
			assert.Equal(t, expected, table.ProviderDefaults)
		})
	}
}

// TestBuildGovernanceDefaultsTableHTTPMiddlewareDefaults 验证 HTTP 中间件默认值稳定性守卫。
func TestBuildGovernanceDefaultsTableHTTPMiddlewareDefaults(t *testing.T) {
	table := BuildGovernanceDefaultsTable(resiliencecontract.GovernanceModeMicroservice)
	require.NotNil(t, table)

	// HTTP 中间件关键默认值
	assert.Equal(t, "15s", table.HTTPMiddlewareDefaults.Timeout)
	assert.Equal(t, "2MB", table.HTTPMiddlewareDefaults.BodyLimit)
	expectedMaxConcurrent := runtime.GOMAXPROCS(0) * 100
	assert.Equal(t, expectedMaxConcurrent, table.HTTPMiddlewareDefaults.MaxConcurrent)
	assert.True(t, table.HTTPMiddlewareDefaults.EnableMetrics)
	assert.False(t, table.HTTPMiddlewareDefaults.EnableCompression)
}

// TestBuildGovernanceDefaultsTableCORSDefaults 验证 CORS 默认值。
func TestBuildGovernanceDefaultsTableCORSDefaults(t *testing.T) {
	table := BuildGovernanceDefaultsTable(resiliencecontract.GovernanceModeMicroservice)
	require.NotNil(t, table)

	cors := table.HTTPMiddlewareDefaults.CORS
	assert.Equal(t, []string{"*"}, cors.AllowOrigins)
	assert.Equal(t, 600, cors.MaxAgeSeconds)
}

// TestBuildGovernanceDefaultsTableSecurityHeaderDefaults 验证安全头默认值。
func TestBuildGovernanceDefaultsTableSecurityHeaderDefaults(t *testing.T) {
	table := BuildGovernanceDefaultsTable(resiliencecontract.GovernanceModeMicroservice)
	require.NotNil(t, table)

	sec := table.HTTPMiddlewareDefaults.SecurityHeaders
	assert.Equal(t, "DENY", sec.XFrameOptions)
	assert.Equal(t, "nosniff", sec.XContentTypeOptions)
	assert.Equal(t, "strict-origin-when-cross-origin", sec.ReferrerPolicy)
}

// TestBuildGovernanceDefaultsTableLocaleDefaults 验证本地化默认值。
func TestBuildGovernanceDefaultsTableLocaleDefaults(t *testing.T) {
	table := BuildGovernanceDefaultsTable(resiliencecontract.GovernanceModeMicroservice)
	require.NotNil(t, table)

	loc := table.HTTPMiddlewareDefaults.Locale
	assert.Equal(t, []string{"zh", "en"}, loc.Supported)
	assert.Equal(t, "zh", loc.Default)
	assert.Equal(t, []string{"lang", "locale"}, loc.QueryKeys)
}

// TestFormatGovernanceDefaultsDiagnosticOutput 验证文本输出包含所有分区和关键值。
func TestFormatGovernanceDefaultsDiagnosticOutput(t *testing.T) {
	summary := GovernanceSummary{
		Mode: resiliencecontract.GovernanceModeMicroservice,
	}
	// 模拟 view=defaults 时的懒加载填充
	summary.Defaults = BuildGovernanceDefaultsTable(resiliencecontract.GovernanceModeMicroservice)

	text := FormatGovernanceDiagnosticView(summary, "defaults")
	require.NotEmpty(t, text)

	// 验证各分区标题存在
	assert.True(t, strings.Contains(text, "Governance Defaults"), "should contain title")
	assert.True(t, strings.Contains(text, "Feature Defaults"), "should contain feature section")
	assert.True(t, strings.Contains(text, "Provider Defaults"), "should contain provider section")
	assert.True(t, strings.Contains(text, "HTTP Middleware Defaults"), "should contain HTTP section")
	assert.True(t, strings.Contains(text, "RPC Client Defaults"), "should contain RPC section")

	// 验证关键默认值可见
	assert.True(t, strings.Contains(text, "timeout: 15s"), "should show timeout default")
	assert.True(t, strings.Contains(text, "body_limit: 2MB"), "should show body_limit default")
	assert.True(t, strings.Contains(text, "x_frame_options: DENY"), "should show security header default")
	assert.True(t, strings.Contains(text, "allow_origins: *"), "should show CORS default")
	assert.True(t, strings.Contains(text, "supported: zh, en"), "should show locale supported")
}
