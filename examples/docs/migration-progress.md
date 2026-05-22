# nopCommerce 微服务重构实施文档

## 项目概述

将 nopCommerce 4.90.4（C# 电商系统）重构为基于 gorp 框架的 Go 微服务架构，使用 multi-independent 模板，输出至 examples/nop-go 目录。

## 已完成工作

### 1. 框架 Bug 修复

#### BUG-001: ginContext.Value 栈溢出修复

**问题描述**：当 ginContext 被包装在 context.WithValue 链中时，调用 Value() 方法会导致无限递归栈溢出。

**修复方案**：
```go
func (c *ginContext) Value(key any) any {
    // 1. 先检查 gin.Context 内部存储
    if strKey, ok := key.(string); ok {
        if val, exists := c.gin.Get(strKey); exists {
            return val
        }
    }
    // 2. 再检查 Request.Context()，避免递归
    return c.gin.Request.Context().Value(key)
}
```

**修复文件**：
- `framework/provider/gin/context.go`
- `framework/provider/gin/context_test.go`（新增测试）

**验证结果**：5 个测试用例全部通过

### 2. 项目结构生成

已生成 20 个微服务骨架：
- gateway-service (8000) - API 网关
- admin-service (8001) - 后台管理
- customer-service (8002) - 客户服务 ✅ 编译通过
- catalog-service (8003) - 商品服务
- cart-service (8004) - 购物车服务
- order-service (8005) - 订单服务
- cms-service (8007) - 内容管理
- inventory-service (8008) - 库存服务
- notification-service (8009) - 通知服务
- payment-service (8010) - 支付服务
- price-service (8011) - 价格服务
- shipping-service (8012) - 物流服务
- store-service (8013) - 多店铺服务
- media-service (8014) - 媒体服务
- localization-service (8015) - 本地化服务
- affiliate-service (8016) - 联盟营销
- seo-service (8017) - SEO 服务
- import-service (8018) - 导入导出
- theme-service (8019) - 主题服务
- ai-service (8020) - AI 服务

### 3. 数据库配置更新

已更新核心服务配置文件，支持环境变量：
- customer-service/config/config.yaml ✅
- catalog-service/config/config.yaml ✅
- order-service/config/config.yaml ✅
- inventory-service/config/config.yaml ✅
- payment-service/config/config.yaml ✅

数据库连接信息：
- 地址：192.168.3.250:3306
- 用户名：admin
- 密码：123456
- 数据库：19 个独立数据库

### 4. 共享模块修复

已修复以下文件：
- `shared/errors/errors.go` - 重写解决编码问题
- `shared/bootstrap/bootstrap.go` - 修复 MustMakeEngine → MustMakeHTTP
- `shared/plugin/adapter.go` - 新增插件适配器
- `shared/plugins/payment-alipay/plugin.go` - 重写支付插件
- `shared/plugins/payment-wechat/plugin.go` - 重写支付插件

### 5. customer-service 完整修复

已修复并编译通过：
- `cmd/main.go` - 修复 BootHTTPService API 使用
- `internal/service/handler.go` - 修复中间件适配

## 待完成工作

### 优先级 P0（立即处理）

1. **批量修复其他服务的 main.go**
   - 所有服务使用相同的模式调用 BootHTTPService
   - 需要将 `bootstrap.BootHTTPService("service-name", opts, migrate, setup)` 改为 `bootstrap.BootHTTPService(opts, migrate, setup)`
   - 需要使用 `container.MustMakeHTTP(rt.Container)` 获取 HTTP 服务

2. **批量修复其他服务的 handler.go**
   - 使用 `ginprovider.AdaptMiddleware()` 适配框架中间件

### 优先级 P1（本周处理）

3. **数据访问层实现**
   - 为每个服务实现 Repository 接口
   - 实现 GORM 模型映射

4. **业务逻辑层实现**
   - 实现 UseCase 层
   - 实现服务间调用逻辑

5. **测试编写**
   - 单元测试（覆盖率 ≥ 80%）
   - 集成测试
   - 端到端测试

### 优先级 P2（后续处理）

6. **容器化部署**
   - Docker Compose 配置验证
   - Kubernetes 部署配置

7. **文档完善**
   - API 文档（Swagger）
   - 架构设计文档
   - 部署文档

## 技术决策记录

### 服务间通信
- 同步调用：gRPC（高性能、强类型）
- 异步通信：Redis Stream（简单可靠）
- 分布式事务：DTM Saga（支持补偿）

### 认证授权
- 用户认证：JWT（框架内置 auth.jwt）
- 服务间认证：Token（框架内置 serviceauth）

### 数据一致性
- 单服务事务：GORM 事务
- 跨服务事务：DTM Saga + Outbox 模式

## 验证命令

```bash
# 编译 shared 模块
cd examples/nop-go && go build ./shared/...

# 编译 customer-service
cd examples/nop-go && go build ./services/customer-service/...

# 运行测试
cd framework/provider/gin && go test -v -run TestGinContextValue
```

## 文件变更清单

### 新增文件
- `framework/provider/gin/context_test.go` - BUG-001 测试用例
- `examples/nop-go/shared/plugin/adapter.go` - 插件适配器
- `examples/nop-go/scripts/update-db-config.sh` - 配置更新脚本

### 修改文件
- `framework/provider/gin/context.go` - BUG-001 修复
- `hade/docs/design/bug-list.md` - Bug 清单更新
- `examples/nop-go/.env.example` - 环境变量模板
- `examples/nop-go/services/customer-service/config/config.yaml` - 数据库配置
- `examples/nop-go/services/catalog-service/config/config.yaml` - 数据库配置
- `examples/nop-go/services/order-service/config/config.yaml` - 数据库配置
- `examples/nop-go/services/inventory-service/config/config.yaml` - 数据库配置
- `examples/nop-go/services/payment-service/config/config.yaml` - 数据库配置
- `examples/nop-go/shared/errors/errors.go` - 重写
- `examples/nop-go/shared/bootstrap/bootstrap.go` - API 修复
- `examples/nop-go/shared/plugins/payment-alipay/plugin.go` - 重写
- `examples/nop-go/shared/plugins/payment-wechat/plugin.go` - 重写
- `examples/nop-go/services/customer-service/cmd/main.go` - API 修复
- `examples/nop-go/services/customer-service/internal/service/handler.go` - 中间件适配

## 下一步行动

1. 使用 Agent 工具批量修复剩余 19 个服务的 main.go 和 handler.go
2. 为每个服务实现数据访问层和业务逻辑层
3. 编写测试用例
4. 验证 Docker Compose 部署