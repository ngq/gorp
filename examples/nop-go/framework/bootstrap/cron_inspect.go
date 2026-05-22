// Package bootstrap provides framework bootstrap and assembly helpers for gorp.
// This file exposes /debug/cron HTTP endpoint for inspecting registered cron jobs.
//
// Bootstrap 包提供 gorp 框架的启动装配辅助能力。
// 本文件暴露 /debug/cron HTTP 端点，用于查看已注册的 cron 任务状态。
package bootstrap

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
	transportcontract "github.com/ngq/gorp/framework/contract/transport"
)

// RegisterCronInspectEndpoints registers the /debug/cron and /doctor/cron endpoints.
// Displays all registered cron jobs with their schedule, last/next run time, and status.
//
// RegisterCronInspectEndpoints 注册 /debug/cron 和 /doctor/cron 端点。
// 展示所有已注册 cron 任务的调度表达式、上次/下次执行时间及执行状态。
func RegisterCronInspectEndpoints(router transportcontract.Router, cronSvc runtimecontract.Cron) {
	if router == nil || cronSvc == nil {
		return
	}

	handler := func(c transportcontract.Context) {
		jobs := cronSvc.Jobs()

		if wantsCronDiagnosticText(c) {
			c.SetHeader("Content-Type", "text/plain; charset=utf-8")
			c.String(http.StatusOK, formatCronDiagnosticView(jobs))
			return
		}
		c.JSON(http.StatusOK, map[string]any{
			"total": len(jobs),
			"jobs":  jobs,
		})
	}

	router.GET("/debug/cron", handler)
	router.GET("/doctor/cron", handler)
}

// wantsCronDiagnosticText checks if the client wants a text/plain response.
//
// wantsCronDiagnosticText 检查客户端是否希望 text/plain 响应。
func wantsCronDiagnosticText(c transportcontract.Context) bool {
	if c == nil {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(c.DefaultQuery("format", "")), "text") {
		return true
	}
	accept := strings.ToLower(strings.TrimSpace(c.GetHeader("Accept")))
	return strings.Contains(accept, "text/plain")
}

// formatCronDiagnosticView formats cron job entries as a human-readable text table.
//
// formatCronDiagnosticView 将 cron 任务条目格式化为可读的文本表格。
func formatCronDiagnosticView(jobs []runtimecontract.CronJobEntry) string {
	if len(jobs) == 0 {
		return "No cron jobs registered.\n"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Cron Jobs (%d registered)\n", len(jobs)))
	sb.WriteString(strings.Repeat("=", 100))
	sb.WriteString("\n")

	for _, j := range jobs {
		name := j.Name
		if name == "" {
			name = "(anonymous)"
		}

		sb.WriteString(fmt.Sprintf("  [%d] %s\n", j.ID, name))
		sb.WriteString(fmt.Sprintf("       Spec:       %s\n", j.Spec))
		sb.WriteString(fmt.Sprintf("       Status:     %s\n", j.Status))

		if !j.LastRunTime.IsZero() {
			sb.WriteString(fmt.Sprintf("       Last Run:   %s\n", j.LastRunTime.Format(time.RFC3339)))
			if j.LastDuration > 0 {
				sb.WriteString(fmt.Sprintf("       Duration:   %s\n", j.LastDuration))
			}
			if j.LastError != "" {
				sb.WriteString(fmt.Sprintf("       Last Error: %s\n", j.LastError))
			}
		} else {
			sb.WriteString("       Last Run:   (never)\n")
		}

		if !j.NextRunTime.IsZero() {
			sb.WriteString(fmt.Sprintf("       Next Run:   %s\n", j.NextRunTime.Format(time.RFC3339)))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
