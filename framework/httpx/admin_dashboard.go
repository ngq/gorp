// Package httpx provides runtime debug admin dashboard endpoints for gorp framework.
package httpx

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	datacontract "github.com/ngq/gorp/framework/contract/data"
	runtimecontract "github.com/ngq/gorp/framework/contract/runtime"
)

var startTime = time.Now()

// DashboardHTML returns the self-contained HTML page for the /debug/gorp Web Console.
func DashboardHTML(prefix string) string {
	if prefix == "" {
		prefix = "/debug/gorp"
	}
	prefix = strings.TrimSuffix(prefix, "/")

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>gorp Developer Console</title>
  <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
  <style>
    body { background-color: #f8f9fa; font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; }
    .navbar-brand { font-weight: 700; color: #0d6efd !important; }
    .card { border-radius: 8px; border: 1px solid #e3e6f0; margin-bottom: 20px; box-shadow: 0 0.15rem 1.75rem 0 rgba(58,59,69,0.15); }
    .card-header { background-color: #f8f9fc; border-bottom: 1px solid #e3e6f0; font-weight: 600; }
    .nav-tabs .nav-link.active { font-weight: 600; border-bottom: 3px solid #0d6efd; color: #0d6efd; }
    .badge-provider { font-size: 0.85rem; padding: 0.4em 0.6em; }
    pre { background-color: #272822; color: #f8f8f2; padding: 15px; border-radius: 6px; max-height: 500px; }
  </style>
</head>
<body>
  <nav class="navbar navbar-expand-lg navbar-dark bg-dark">
    <div class="container-fluid">
      <a class="navbar-brand ms-3" href="%s">🚀 gorp Developer Console</a>
      <span class="navbar-text me-3 text-light" id="system-uptime">Uptime: Loading...</span>
    </div>
  </nav>

  <div class="container-fluid mt-4 px-4">
    <ul class="nav nav-tabs mb-4" id="dashboardTabs" role="tablist">
      <li class="nav-item"><button class="nav-link active" data-bs-toggle="tab" data-bs-target="#container-tab">🧩 DI 容器拓扑 (DAG)</button></li>
      <li class="nav-item"><button class="nav-link" data-bs-toggle="tab" data-bs-target="#routes-tab">🌐 路由与中间件</button></li>
      <li class="nav-item"><button class="nav-link" data-bs-toggle="tab" data-bs-target="#config-tab">⚙️ 动态配置</button></li>
      <li class="nav-item"><button class="nav-link" data-bs-toggle="tab" data-bs-target="#metrics-tab">📈 实时指标与 BBR 熔断</button></li>
      <li class="nav-item"><button class="nav-link" data-bs-toggle="tab" data-bs-target="#swagger-tab">📖 交互式 Swagger API</button></li>
    </ul>

    <div class="tab-content" id="dashboardTabContent">
      <!-- 1. DI Container Tab -->
      <div class="tab-pane fade show active" id="container-tab">
        <div class="card">
          <div class="card-header">已注册 Provider 与依赖 DAG 图</div>
          <div class="card-body">
            <div class="table-responsive">
              <table class="table table-hover align-middle">
                <thead class="table-light">
                  <tr>
                    <th>Provider 名称</th>
                    <th>加载状态</th>
                    <th>引导状态 (Boot)</th>
                    <th>延迟加载 (IsDefer)</th>
                    <th>提供契约 (Provides)</th>
                    <th>依赖契约 (DependsOn)</th>
                  </tr>
                </thead>
                <tbody id="container-tbody">
                  <tr><td colspan="6" class="text-center text-muted">加载容器拓扑数据中...</td></tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <!-- 2. Routes Tab -->
      <div class="tab-pane fade" id="routes-tab">
        <div class="card">
          <div class="card-header">注册 HTTP 路由清单</div>
          <div class="card-body">
            <div class="table-responsive">
              <table class="table table-hover align-middle">
                <thead class="table-light">
                  <tr>
                    <th>HTTP Method</th>
                    <th>路由 Path</th>
                    <th>Handler 函数名</th>
                  </tr>
                </thead>
                <tbody id="routes-tbody">
                  <tr><td colspan="3" class="text-center text-muted">加载路由表中...</td></tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>

      <!-- 3. Config Tab -->
      <div class="tab-pane fade" id="config-tab">
        <div class="card">
          <div class="card-header d-flex justify-content-between align-items-center">
            <span>在线配置视图（自动脱敏处理）</span>
            <button class="btn btn-sm btn-outline-primary" onclick="loadConfig()">刷新配置</button>
          </div>
          <div class="card-body">
            <pre><code id="config-json">加载配置数据中...</code></pre>
          </div>
        </div>
      </div>

      <!-- 4. Metrics Tab -->
      <div class="tab-pane fade" id="metrics-tab">
        <div class="row">
          <div class="col-md-4">
            <div class="card text-center p-3">
              <h6 class="text-muted">Goroutines 活跃数</h6>
              <h2 id="metric-goroutines" class="text-primary">--</h2>
            </div>
          </div>
          <div class="col-md-4">
            <div class="card text-center p-3">
              <h6 class="text-muted">系统 CPU 核心数 / 平台</h6>
              <h2 id="metric-cpus" class="text-success">--</h2>
            </div>
          </div>
          <div class="col-md-4">
            <div class="card text-center p-3">
              <h6 class="text-muted">内存占用 (HeapAlloc)</h6>
              <h2 id="metric-memory" class="text-info">-- MB</h2>
            </div>
          </div>
        </div>
      </div>

      <!-- 5. Swagger UI Tab -->
      <div class="tab-pane fade" id="swagger-tab">
        <iframe id="swagger-iframe" src="%s/swagger" style="width:100%%; height:750px; border:none;"></iframe>
      </div>
    </div>
  </div>

  <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/js/bootstrap.bundle.min.js"></script>
  <script>
    const prefix = "%s";

    function fetchJSON(url, callback) {
      fetch(url).then(r => r.json()).then(data => callback(data)).catch(e => console.error(e));
    }

    function loadContainer() {
      fetchJSON(prefix + "/api/container", data => {
        const tbody = document.getElementById("container-tbody");
        if (!data || data.length === 0) {
          tbody.innerHTML = '<tr><td colspan="6" class="text-center text-muted">暂无 Provider 数据</td></tr>';
          return;
        }
        let html = "";
        data.forEach(p => {
          html += '<tr>' +
            '<td><strong>' + p.Name + '</strong></td>' +
            '<td>' + (p.Loaded ? '<span class="badge bg-success">Loaded</span>' : '<span class="badge bg-secondary">Unloaded</span>') + '</td>' +
            '<td>' + (p.Booted ? '<span class="badge bg-primary">Booted</span>' : '<span class="badge bg-light text-dark">Idle</span>') + '</td>' +
            '<td>' + (p.IsDefer ? '<span class="badge bg-warning text-dark">Defer</span>' : '<span class="badge bg-info">Eager</span>') + '</td>' +
            '<td><code>' + (p.Provides ? p.Provides.join(", ") : "-") + '</code></td>' +
            '<td><code>' + (p.DependsOn ? p.DependsOn.join(", ") : "-") + '</code></td>' +
            '</tr>';
        });
        tbody.innerHTML = html;
      });
    }

    function loadRoutes() {
      fetchJSON(prefix + "/api/routes", data => {
        const tbody = document.getElementById("routes-tbody");
        if (!data || data.length === 0) {
          tbody.innerHTML = '<tr><td colspan="3" class="text-center text-muted">暂无路由数据</td></tr>';
          return;
        }
        let html = "";
        data.forEach(r => {
          let badgeClass = "bg-primary";
          if (r.method === "POST") badgeClass = "bg-success";
          if (r.method === "PUT") badgeClass = "bg-warning text-dark";
          if (r.method === "DELETE") badgeClass = "bg-danger";
          html += '<tr>' +
            '<td><span class="badge ' + badgeClass + '">' + r.method + '</span></td>' +
            '<td><code>' + r.path + '</code></td>' +
            '<td><small class="text-muted">' + r.handler + '</small></td>' +
            '</tr>';
        });
        tbody.innerHTML = html;
      });
    }

    function loadConfig() {
      fetchJSON(prefix + "/api/config", data => {
        document.getElementById("config-json").textContent = JSON.stringify(data, null, 2);
      });
    }

    function loadMetrics() {
      fetchJSON(prefix + "/api/metrics", data => {
        document.getElementById("system-uptime").textContent = "Uptime: " + data.uptime;
        document.getElementById("metric-goroutines").textContent = data.goroutines;
        document.getElementById("metric-cpus").textContent = data.num_cpu + " Cores / " + data.goos;
        document.getElementById("metric-memory").textContent = (data.alloc_bytes / (1024 * 1024)).toFixed(2) + " MB";
      });
    }

    document.addEventListener("DOMContentLoaded", () => {
      loadContainer();
      loadRoutes();
      loadConfig();
      loadMetrics();
      setInterval(loadMetrics, 3000);
    });
  </script>
</body>
</html>`, prefix, prefix, prefix)
}

// MountDebugDashboard attaches the /debug/gorp developer console and JSON API endpoints to Gin.
func MountDebugDashboard(engine *gin.Engine, container runtimecontract.Container, prefix ...string) {
	if engine == nil {
		return
	}
	pathPrefix := "/debug/gorp"
	if len(prefix) > 0 && prefix[0] != "" {
		pathPrefix = prefix[0]
	}
	pathPrefix = strings.TrimSuffix(pathPrefix, "/")

	// 1. Dashboard Main HTML
	engine.GET(pathPrefix, func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(DashboardHTML(pathPrefix)))
	})
	engine.GET(pathPrefix+"/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(DashboardHTML(pathPrefix)))
	})

	// 2. Swagger UI Sub-route
	swaggerHandler := GinSwaggerHandler("")
	engine.GET(pathPrefix+"/swagger", func(c *gin.Context) {
		// Auto-generate OpenAPI 3 spec on the fly for Swagger UI
		reqPath := c.Request.URL.Path
		if strings.HasSuffix(reqPath, "/spec") || strings.HasSuffix(reqPath, ".json") {
			specJSON := ExportOpenAPI3JSON(engine, "gorp Service API Spec", "1.0.0")
			c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(specJSON))
			return
		}
		swaggerHandler(c)
	})
	engine.GET(pathPrefix+"/swagger/*any", func(c *gin.Context) {
		reqPath := c.Request.URL.Path
		if strings.HasSuffix(reqPath, "/spec") || strings.HasSuffix(reqPath, ".json") {
			specJSON := ExportOpenAPI3JSON(engine, "gorp Service API Spec", "1.0.0")
			c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(specJSON))
			return
		}
		swaggerHandler(c)
	})

	// 3. API - Container DAG Topology
	engine.GET(pathPrefix+"/api/container", func(c *gin.Context) {
		if container == nil {
			c.JSON(http.StatusOK, []any{})
			return
		}
		dag := container.ProviderDAG()
		c.JSON(http.StatusOK, dag.Nodes)
	})

	// 4. API - Routes
	engine.GET(pathPrefix+"/api/routes", func(c *gin.Context) {
		routes := engine.Routes()
		res := make([]map[string]string, 0, len(routes))
		for _, r := range routes {
			if strings.HasPrefix(r.Path, pathPrefix) {
				continue
			}
			res = append(res, map[string]string{
				"method":  r.Method,
				"path":    r.Path,
				"handler": r.Handler,
			})
		}
		c.JSON(http.StatusOK, res)
	})

	// 5. API - Config (With Secret Masking)
	engine.GET(pathPrefix+"/api/config", func(c *gin.Context) {
		if container == nil || !container.IsBind(datacontract.ConfigKey) {
			c.JSON(http.StatusOK, map[string]any{"status": "no config provider bound"})
			return
		}
		cfgAny, err := container.Make(datacontract.ConfigKey)
		if err != nil {
			c.JSON(http.StatusOK, map[string]any{"error": err.Error()})
			return
		}
		if cfg, ok := cfgAny.(datacontract.Config); ok {
			var settings map[string]any
			if sp, ok := cfg.(interface{ AllSettings() map[string]any }); ok {
				settings = sp.AllSettings()
			} else {
				_ = cfg.Unmarshal("", &settings)
			}
			c.JSON(http.StatusOK, maskSecrets(settings))
			return
		}
		c.JSON(http.StatusOK, map[string]any{"status": "config provider loaded"})
	})

	// 6. API - Metrics
	engine.GET(pathPrefix+"/api/metrics", func(c *gin.Context) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		c.JSON(http.StatusOK, map[string]any{
			"uptime":      time.Since(startTime).Round(time.Second).String(),
			"goroutines":  runtime.NumGoroutine(),
			"num_cpu":     runtime.NumCPU(),
			"goos":        runtime.GOOS,
			"alloc_bytes": m.Alloc,
			"sys_bytes":   m.Sys,
		})
	})

	// 7. API - Auto-Exported OpenAPI Spec JSON
	engine.GET(pathPrefix+"/openapi.json", func(c *gin.Context) {
		specJSON := ExportOpenAPI3JSON(engine, "gorp Auto-Exported API Spec", "1.0.0")
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(specJSON))
	})
}

func maskSecrets(m map[string]any) map[string]any {
	res := make(map[string]any)
	for k, v := range m {
		lowerKey := strings.ToLower(k)
		if strings.Contains(lowerKey, "password") || strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "token") || strings.Contains(lowerKey, "key") {
			res[k] = "****** (redacted)"
		} else if nestedMap, ok := v.(map[string]any); ok {
			res[k] = maskSecrets(nestedMap)
		} else {
			res[k] = v
		}
	}
	return res
}
