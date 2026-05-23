# Middleware

HTTP 中间件预留目录。

当前 catalog-service 使用 gorp 框架内置的中间件（日志、链路追踪、认证等）。
业务自定义中间件可在此目录扩展，例如：
- 商品访问频率限制
- 媒体上传大小校验
- SEO 数据脱敏
