# HTTP Middleware

admin-service 的自定义中间件目前为占位实现。

如需添加管理后台专属中间件（如操作审计、权限校验、租户隔离等），
可在 `internal/server/http/middleware/` 下添加。

框架级通用中间件（链路追踪、限流、熔断等）由 gorp 治理层统一提供，
无需在此重复实现。