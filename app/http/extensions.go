package http

import (
	"net/http"

	frameworkcontainer "github.com/ngq/gorp/framework/container"
	"github.com/ngq/gorp/framework/contract"
	ginprovider "github.com/ngq/gorp/framework/provider/gin"

	"github.com/gin-gonic/gin"
)

// RegisterExtensionDemoRoutes 注册一组“框架能力探针”示例路由。
//
// 中文说明：
// - 这些路由不属于正式业务接口，而是用于演示/验证扩展能力是否已经接通。
// - 当前覆盖：gorm、sqlx、redis、cron 等基础组件。
func RegisterExtensionDemoRoutes(c contract.Container) error {
	engine, err := ginprovider.EngineFromContainer(c)
	if err != nil {
		return err
	}

	engine.GET("/api/ext/db/gorm", func(ctx *gin.Context) {
		db, err := frameworkcontainer.MakeGormDB(c)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		sqlDB, err := db.DB()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := sqlDB.PingContext(ctx.Request.Context()); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	engine.GET("/api/ext/db/sqlx", func(ctx *gin.Context) {
		db, err := frameworkcontainer.MakeSQLX(c)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := db.PingContext(ctx.Request.Context()); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	engine.GET("/api/ext/redis/ping", func(ctx *gin.Context) {
		r, err := frameworkcontainer.MakeRedis(c)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := r.Ping(ctx.Request.Context()); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	engine.GET("/api/ext/cron/start", func(ctx *gin.Context) {
		cr, err := frameworkcontainer.MakeCron(c)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		cr.Start()
		ctx.JSON(http.StatusOK, gin.H{"status": "started"})
	})

	return nil
}
