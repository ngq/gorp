// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file exposes formal inspect/doctor HTTP view for governance results.
// Reuses the same governance summary used by startup logs and tests.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件为最终治理生效结果暴露统一的 inspect/doctor HTTP 视图。
// 复用启动日志和测试已经使用的同一份治理摘要。
package bootstrap

import (
	"net/http"
	"strings"

	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// RegisterGovernanceInspectEndpoints registers the formal governance inspect/doctor endpoints.
//
// RegisterGovernanceInspectEndpoints 注册正式治理 inspect / doctor 端点。
func RegisterGovernanceInspectEndpoints(router transportcontract.Router, summary GovernanceSummary) {
	if router == nil {
		return
	}

	handler := func(c transportcontract.Context) {
		view := strings.ToLower(strings.TrimSpace(c.DefaultQuery("view", "")))

		// view=defaults 时懒加载默认值表，不影响共享 summary
		if view == "defaults" {
			summaryCopy := summary
			summaryCopy.Defaults = BuildGovernanceDefaultsTable(summary.Mode)
			if wantsGovernanceDiagnosticText(c) {
				c.SetHeader("Content-Type", "text/plain; charset=utf-8")
				c.String(http.StatusOK, FormatGovernanceDiagnosticView(summaryCopy, view))
				return
			}
			c.JSON(http.StatusOK, summaryCopy)
			return
		}

		if wantsGovernanceDiagnosticText(c) {
			c.SetHeader("Content-Type", "text/plain; charset=utf-8")
			c.String(http.StatusOK, FormatGovernanceDiagnosticView(summary, c.DefaultQuery("view", "")))
			return
		}
		c.JSON(http.StatusOK, summary)
	}

	router.GET("/debug/governance", handler)
	router.GET("/doctor/governance", handler)
}

func wantsGovernanceDiagnosticText(c transportcontract.Context) bool {
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
