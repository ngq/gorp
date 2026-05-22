# shared

这里只放跨服务稳定复用的公共内容。

当前默认保留三类最小共享 helper：

- `shared/config/`：通用配置加载 helper
- `shared/db/`：通用数据库打开 helper
- `shared/logger/`：通用日志门面 helper

不应该放这里的内容：

- 业务逻辑
- 服务私有装配
- 中间件接入细节
- 只被单个服务使用的代码

判断规则很简单：

> 如果它还没有成为两个及以上服务稳定复用的内容，就先留在服务自己的 `internal/` 里。
