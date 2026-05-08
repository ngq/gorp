// Application scenarios:
// - Expose one formal inspect/doctor HTTP view for the effective governance result.
// - Reuse the same governance summary used by startup logs and tests.
// - Make implicit defaults queryable at runtime without introducing a second summary model.
//
// 适用场景：
// - 为最终治理生效结果暴露统一的 inspect / doctor HTTP 视图。
// - 复用启动日志和测试已经使用的同一份治理摘要。
// - 让隐式默认在运行时可查询，而不是再引入第二套摘要模型。
package bootstrap

import (
	"net/http"
	"strings"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// RegisterGovernanceInspectEndpoints registers the formal governance inspect/doctor endpoints.
//
// RegisterGovernanceInspectEndpoints 注册正式治理 inspect / doctor 端点。
func RegisterGovernanceInspectEndpoints(router transportcontract.HTTPRouter, summary GovernanceSummary) {
	if router == nil {
		return
	}

	handler := func(c transportcontract.HTTPContext) {
		if wantsGovernanceDiagnosticText(c) {
			c.Header("Content-Type", "text/plain; charset=utf-8")
			c.String(http.StatusOK, FormatGovernanceDiagnosticView(summary, c.DefaultQuery("view", "")))
			return
		}
		c.JSON(http.StatusOK, summary)
	}

	router.GET("/debug/governance", handler)
	router.GET("/doctor/governance", handler)
}

func wantsGovernanceDiagnosticText(c transportcontract.HTTPContext) bool {
	if c == nil {
		return false
	}
	if strings.TrimSpace(c.DefaultQuery("view", "")) != "" {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(c.DefaultQuery("format", "")), "text") {
		return true
	}
	accept := strings.ToLower(strings.TrimSpace(c.GetHeader("Accept")))
	return strings.Contains(accept, "text/plain")
}
