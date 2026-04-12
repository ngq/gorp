package http

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/ngq/gorp/framework/contract"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"

	"github.com/gin-gonic/gin"
)

// RegisterFrontend 注册前端集成逻辑。
//
// 中文说明：
// - 开发态：如果设置了 FRONTEND_DEV_SERVER，则把所有非 API 路由反向代理到前端 dev server。
// - 生产态：如果 frontend/dist 存在，则直接托管静态资源，并提供 SPA 回退。
// - 这样同一套后端服务可以同时适配“本地联调”和“构建后托管”两种场景。
func RegisterFrontend(c contract.Container) error {
	r, err := ginprovider.EngineFromContainer(c)
	if err != nil {
		return err
	}

	devServer := strings.TrimSpace(os.Getenv("FRONTEND_DEV_SERVER"))
	if devServer != "" {
		target, err := url.Parse(devServer)
		if err != nil {
			return err
		}
		proxy := httputil.NewSingleHostReverseProxy(target)

		// 中文说明：
		// - 前端开发服务器只兜底处理“非后端路由”。
		// - API/user/question/answer/swagger 等路径仍然保留给后端处理。
		r.NoRoute(func(ctx *gin.Context) {
			// Let backend routes handle API paths.
			path := ctx.Request.URL.Path
			if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/user/") || strings.HasPrefix(path, "/question/") || strings.HasPrefix(path, "/answer/") || strings.HasPrefix(path, "/swagger/") {
				ctx.Status(http.StatusNotFound)
				return
			}
			proxy.ServeHTTP(ctx.Writer, ctx.Request)
		})
		return nil
	}

	// Production static hosting
	dist := filepath.Join("frontend", "dist")
	if st, err := os.Stat(dist); err == nil && st.IsDir() {
		r.Static("/", dist)
		// SPA fallback
		r.NoRoute(func(ctx *gin.Context) {
			ctx.File(filepath.Join(dist, "index.html"))
		})
	}

	return nil
}
