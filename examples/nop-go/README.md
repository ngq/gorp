# nop-go

`nop-go` 是 `gorp` 在 `examples/` 下的一套**真实业务风格示例项目**，目标不是最小 demo，而是用一组商城/中后台相关服务，验证：

- framework 主链能力（bootstrap/module_groups、auth.jwt、orm.runtime）
- 多服务拆分后的目录与分层约定
- auth / JWT / gateway / plugin / payment 等横切能力如何落地
- 真实业务团队可直接开发的骨架

> 当前状态：
>
> - **已对齐 framework 主链**：使用 `framework/bootstrap/module_groups` 组合 Provider
> - **认证链路统一**：使用 framework 级 `auth.jwt` 能力，配置从 `auth.jwt.*` 读取
> - **启动装配已统一收口**：各服务 `cmd/main.go` 已改为通过 `wire.go / wire_gen.go` 风格入口收口依赖装配
> - **已完成一轮全量验证**：20 个服务 + `shared` 已通过 `go build ./...` 与 `go test ./...`
> - 不再依赖项目层 `shared/auth` 实现

---

## 1. 目录说明

- `services/`：各业务服务
- `shared/`：共享能力（bootstrap、plugin、rpc、errors 等）
- `infra/proto/`：proto 定义
- `infra/pb/`：proto 生成产物（当前目录已预留）
- `docs/`：`nop-go` 私有/专项文档
- `go.work`：工作区配置，聚合所有服务和 shared 模块

---

## 2. 服务清单

当前工作区包含：

- `catalog-service`：商品目录（商品/分类/品牌）
- `customer-service`：客户、地址、登录
- `inventory-service`：库存与仓库
- `price-service`：价格、税率、折扣
- `order-service`：订单、退货、礼品卡
- `payment-service`：支付与退款
- `shipping-service`：发货与配送方式
- `cart-service`：购物车与愿望清单
- `cms-service`：博客/新闻/论坛/主题
- `notification-service`：通知与模板
- `gateway-service`：统一 HTTP 网关
- `admin-service`：后台用户/角色/设置/日志

---

## 3. Framework 主链对齐

### 3.1 Provider 组装

各服务通过 `shared/bootstrap` 统一启动，底层使用 framework 的 module_groups：

```go
// Foundation: app/config/log/gin/host/cron
bootstrap.FoundationProviders()

// ORM/Runtime: gorm/sqlx/runtime/inspect
bootstrap.ORMRuntimeProviders()

// BusinessSimplification: redis/auth.jwt/serviceauth
bootstrap.BusinessSimplificationProviders()
```

### 3.2 认证配置

使用 framework 级 `auth.jwt` 配置（不再使用 `jwt_secret`/`jwt_expire`）：

```yaml
auth:
  jwt:
    secret: “your-secret-key”
    issuer: “your-service-name”
    audience: “your-audience”
```

### 3.3 JWT 服务使用

```go
// 从容器获取 JWTService
jwtSvc := bootstrap.MustMakeJWTService(rt.Container)

// 签发 token
claims := jwtSvc.NewClaims(userID, “customer”, username, roles, 86400)
token, err := jwtSvc.Sign(claims)

// 使用 framework 中间件
jwtmiddleware.AuthMiddleware(jwtSvc, “customer”)
```

---

## 4. 推荐使用方式

### 4.1 Lightweight 模式（推荐先这样用）

如果你想快速理解项目、验证主链路，建议先聚焦：

1. `catalog-service`
2. `customer-service`
3. `cart-service`
4. `gateway-service`
5. （可选）`admin-service`
6. （可选）`payment-service`

这条路径适合：

- 快速做商品/客户/购物车基本业务验证
- 先验证 shared/bootstrap、JWT、gateway 路由治理
- 避免一开始就把所有服务全跑起来

### 4.2 Full Demo 模式（后续再用）

当你需要验证多服务协作时，再扩到：

- order / inventory / shipping / notification / cms / price / payment 全链路

---

## 5. 当前最短开发路径

### 第一步：进入工作区

```bash
cd examples/nop-go
```

### 第二步：先编译共享模块与关键服务

```bash
cd shared && go build ./...
cd ../services/customer-service && go build ./...
cd ../cart-service && go build ./...
cd ../catalog-service && go build ./...
cd ../gateway-service && go build ./...
```

### 第三步：按需启动单个服务

```bash
cd examples/nop-go
make start-customer-service
make start-catalog-service
make start-cart-service
make start-gateway-service
```

或者手工：

```bash
cd services/customer-service && go run ./cmd
```

### 第四步：检查健康状态

- gateway：`http://localhost:8000/healthz`
- 其他服务：各自端口 `/healthz`

---

## 6. 当前配置与环境约定

### 当前已有约定

各服务使用：

- `config/config.yaml`
- 认证配置使用 `auth.jwt.*` 格式

网关配置中：

- `routes` 已作为当前唯一真相源
- `app.env` 已开始收敛为 `dev`

### 当前建议

- 目前先使用 `dev / test / prod` 作为环境命名方向
- 敏感信息（如数据库 dsn、jwt secret）后续应逐步迁到环境变量覆盖
- 当前仓库里的示例配置仍偏开发态，不应直接照搬到生产

---

## 7. 当前主链路进展

### 已完成收口

- ✅ `shared/bootstrap` 使用 framework `module_groups` 组合
- ✅ admin/customer/cart 服务使用 framework `auth.jwt` 能力
- ✅ JWT 配置统一为 `auth.jwt.*` 格式
- ✅ 删除项目层 `shared/auth` 目录，改用 framework 实现
- ✅ gateway 路由统一到 `config/config.yaml`
- ✅ framework 补充 `MakeHost/MakeHTTP/MakeGinEngine/MustMake*` helper
- ✅ 主要服务的 `cmd/main.go` 已统一收口为 `wire.go / wire_gen.go` 风格启动装配
- ✅ 已完成 20 个服务 + `shared` 的 `go build ./...` 与 `go test ./...` 回归

### 仍在继续收口的部分

- `examples/nop-go` 顶层 README 与 per-service README 仍可继续完善
- 这轮 Wire 收口经验仍可继续回灌到更多模板细节
- 特殊链路（plugin / dlock / gateway 配置行为）的专项说明仍可继续补强

---

## 8. 真实业务团队最该关注的 5 件事

1. **Framework 主链对齐**
   - 所有服务统一使用 `framework/bootstrap/module_groups`

2. **auth / JWT 统一**
   - 使用 framework `auth.jwt` 能力，配置从 `auth.jwt.*` 读取

3. **配置治理**
   - 统一使用 `auth.jwt.*` 格式，不再使用 `jwt_secret`/`jwt_expire`

4. **简单服务轻量模式**
   - 当前以 `catalog-service` 为试点，逐步把非核心子域变成可选装配

5. **文档入口**
   - 本 README 是第一步，后续还需要 per-service README 与架构说明

---

## 9. 当前已知限制

请在使用前明确这些边界：

1. 这不是“已经收口完成的生产模板”
2. 虽然已完成一轮 20 个服务 + `shared` 的 build/test 回归，但多数模块当前仍缺少真正的测试文件（大量包为 `[no test files]`）
3. `Makefile` 中某些目标当前仍是预留态：
   - `init-db` 依赖的 `infra/init_db.go` 当前未落地
4. 本地 Docker 一键部署链路正在继续收口；当前宿主机数据库端口为 MySQL `13306`、Redis `16379`
5. `payment-service` 插件体系有设计和基础实现，但还需要进一步说明与验证
6. 完整多服务协作链路仍建议按模块逐步验证，不要一开始全量启动

---

## 10. 下一步推荐

如果你是：

### 想理解 framework 主链

优先看：

- `shared/bootstrap/bootstrap.go`
- `framework/bootstrap/module_groups.go`
- `framework/provider/auth/jwt/provider.go`

### 想继续收口真实业务主链路

优先看：

- `customer-service`
- `cart-service`
- `admin-service`

### 想做轻量模式

优先看：

- `catalog-service/internal/service/handler.go`
- `catalog-service/cmd/main.go`

### 想看插件化支付方向

优先看：

- `docs/plugin-system-design.zh-CN.md`
- `shared/plugin/`
- `shared/plugins/payment-*`

---

## 11. 一句话总结

`examples/nop-go` 当前最适合被理解为：

**”已完成 framework 主链对齐，并完成一轮全量 build/test 回归的真实业务风格样例”**，可直接作为多服务项目骨架参考。
