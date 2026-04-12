package gorm

import (
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"gorm.io/gorm"
)

var (
	// dbConnectionsOpen 当前打开的数据库连接数
	dbConnectionsOpen = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gorp_db_connections_open",
		Help: "The number of established database connections (in-use and idle).",
	}, []string{"driver"})

	// dbConnectionsInUse 正在使用的数据库连接数
	dbConnectionsInUse = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gorp_db_connections_in_use",
		Help: "The number of database connections currently in use.",
	}, []string{"driver"})

	// dbConnectionsIdle 空闲的数据库连接数
	dbConnectionsIdle = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gorp_db_connections_idle",
		Help: "The number of idle database connections.",
	}, []string{"driver"})

	// dbConnectionsWaitTotal 等待连接的总次数
	dbConnectionsWaitTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_db_connections_wait_total",
		Help: "The total number of connections waited for.",
	}, []string{"driver"})

	// dbConnectionsWaitDuration 等待连接的总耗时
	dbConnectionsWaitDuration = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gorp_db_connections_wait_duration_seconds",
		Help: "The total time waited for connections.",
	}, []string{"driver"})

	// dbQueryDuration 数据库查询耗时
	dbQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gorp_db_query_duration_seconds",
		Help:    "Database query latency in seconds.",
		Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"driver", "operation"})
)

// DBMetricsCollector 收集数据库连接池指标。
//
// 中文说明：
// - 定期从 sql.DBStats 获取连接池状态并更新 Prometheus 指标；
// - 用于监控数据库连接池健康状态，排查连接泄漏等问题；
// - 建议在应用启动后调用 StartCollection 开始定期收集。
type DBMetricsCollector struct {
	sqlDB   *sql.DB
	driver  string
	stopCh  chan struct{}
	ticker  *time.Ticker
}

// NewDBMetricsCollector 创建数据库指标收集器。
func NewDBMetricsCollector(sqlDB *sql.DB, driver string) *DBMetricsCollector {
	return &DBMetricsCollector{
		sqlDB:  sqlDB,
		driver: driver,
		stopCh: make(chan struct{}),
	}
}

// StartCollection 开始定期收集数据库连接池指标。
//
// 中文说明：
// - 每 5 秒采集一次 sql.DBStats 数据；
// - 更新 Prometheus Gauge 指标；
// - 返回 stop 函数，调用后停止收集。
func (c *DBMetricsCollector) StartCollection() func() {
	c.ticker = time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-c.ticker.C:
				c.collect()
			case <-c.stopCh:
				c.ticker.Stop()
				return
			}
		}
	}()
	return c.Stop
}

// Stop 停止收集指标。
func (c *DBMetricsCollector) Stop() {
	close(c.stopCh)
}

// collect 采集当前连接池状态。
func (c *DBMetricsCollector) collect() {
	stats := c.sqlDB.Stats()

	dbConnectionsOpen.WithLabelValues(c.driver).Set(float64(stats.OpenConnections))
	dbConnectionsInUse.WithLabelValues(c.driver).Set(float64(stats.InUse))
	dbConnectionsIdle.WithLabelValues(c.driver).Set(float64(stats.Idle))
	dbConnectionsWaitTotal.WithLabelValues(c.driver).Add(float64(stats.WaitCount))
	dbConnectionsWaitDuration.WithLabelValues(c.driver).Add(stats.WaitDuration.Seconds())
}

// GormQueryCallback 为 GORM 添加查询耗时回调。
//
// 中文说明：
// - 通过 GORM 的 Callback 机制记录每次查询的耗时；
// - 区分 create/query/update/delete 四种操作类型；
// - 用于监控数据库查询性能，识别慢查询。
func GormQueryCallback(db *gorm.DB, driver string) {
	// 注册 before callback
	db.Callback().Create().Before("gorm:create").Register("prometheus:before_create", func(d *gorm.DB) {
		d.Set("prometheus:start_time", time.Now())
	})
	db.Callback().Query().Before("gorm:query").Register("prometheus:before_query", func(d *gorm.DB) {
		d.Set("prometheus:start_time", time.Now())
	})
	db.Callback().Update().Before("gorm:update").Register("prometheus:before_update", func(d *gorm.DB) {
		d.Set("prometheus:start_time", time.Now())
	})
	db.Callback().Delete().Before("gorm:delete").Register("prometheus:before_delete", func(d *gorm.DB) {
		d.Set("prometheus:start_time", time.Now())
	})

	// 注册 after callback
	db.Callback().Create().After("gorm:create").Register("prometheus:after_create", func(d *gorm.DB) {
		recordQueryDuration(d, driver, "create")
	})
	db.Callback().Query().After("gorm:query").Register("prometheus:after_query", func(d *gorm.DB) {
		recordQueryDuration(d, driver, "query")
	})
	db.Callback().Update().After("gorm:update").Register("prometheus:after_update", func(d *gorm.DB) {
		recordQueryDuration(d, driver, "update")
	})
	db.Callback().Delete().After("gorm:delete").Register("prometheus:after_delete", func(d *gorm.DB) {
		recordQueryDuration(d, driver, "delete")
	})
}

// recordQueryDuration 记录查询耗时。
func recordQueryDuration(d *gorm.DB, driver, operation string) {
	startTime, ok := d.Get("prometheus:start_time")
	if !ok {
		return
	}
	if t, ok := startTime.(time.Time); ok {
		duration := time.Since(t).Seconds()
		dbQueryDuration.WithLabelValues(driver, operation).Observe(duration)
	}
}