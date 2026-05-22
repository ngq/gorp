# gRPC 开发指南

本文档详细说明：

1. 如何使用 `gorp` 命令创建项目模板
2. 框架中 gRPC 的两种使用方式及其选择建议
3. 基于 Proto-first 的最佳实践

---

## 第一部分：使用 gorp CLI 创建项目

## 安装 gorp 命令

```bash
# 从 GitHub 源码安装
go install github.com/ngq/gorp/cmd/gorp@latest

# 或者安装特定版本
# go install github.com/ngq/gorp/cmd/gorp@v0.1.0
```

安装后验证：

```bash
gorp --help
```

**注意**：如果 GitHub 仓库尚未发布，可以临时使用本地源码：

```bash
# 本地开发环境临时安装
cd D:\project\gorp
go install ./cmd/gorp
```

## gorp 命令概览

```text
gorp - Framework, starter templates, and developer tooling for gorp

一级命令：
  new          创建新项目（推荐入口）
  template     模板治理与资产维护
  proto        Proto 文件生成
  model        模型脚手架
  app          应用运行时管理
  grpc         gRPC 服务管理
  cron         定时任务管理
  build        构建命令
  dev          开发辅助
  deploy       部署命令
  doc          文档生成
  swagger      Swagger 文档
  provider     Provider 脚手架
  middleware   Middleware 脚手架
  command      Command 脚手架
```

## gorp new 命令详解

### 基本用法

```bash
# 默认单服务快速启动（golayout 模板）
gorp new

# 单服务 + Wire 依赖注入
gorp new wire

# 主推微服务模板（多服务 + Wire + Proto）
gorp new multi-wire
```

### 可用模板

| 模板名 | 说明 | 适用场景 | gRPC 支持 |
|-------|------|---------|----------|
| `golayout` | 默认单服务模板 | 快速启动单体项目 | ❌ |
| `golayout-wire` | 单服务 Wire 高级模板 | 需要显式编译期装配的单服务项目 | ❌ |
| `multi-flat-wire` | 主推微服务模板（多服务 + Wire + Proto） | 当前公开微服务主线 | ✅ |
| `multi-independent` | 公开进阶多服务模板 | 更强服务自治、独立 go.mod、后续拆仓 | 视项目接入而定 |
| `base` | 最小骨架 | 自定义结构 | ❌ |

### 命令参数

```bash
gorp new [intent] [flags]

# intent（位置参数）
  wire        -> golayout-wire 模板
  multi-wire  -> multi-flat-wire 模板（推荐用于 gRPC）

# 进阶多服务模板
  --template=multi-independent

# flags
  --template   指定模板名（优先级高于 intent）
  --backend    数据库后端：gorm|ent（默认 gorm）
  --with-db    是否包含数据库示例（默认 true）
  --with-swagger 是否启用 Swagger（默认 true）
```

### 创建项目示例

```bash
# 示例 1：创建带 gRPC 的多服务项目（推荐）
gorp new multi-wire
# 输入：目录名称=my-microservice
# 输入：模块名=github.com/myorg/my-microservice
# 输入：框架路径=D:\project\gorp

# 示例 2：创建单服务项目
gorp new

# 示例 3：使用 ent 后端
gorp new multi-wire --backend=ent

# 示例 4：无数据库的最小模板
gorp new --template=base --with-db=false
```

### 创建过程示例

```bash
$ gorp new multi-wire
请输入目录名称：my-microservice
请输入模块名称(go.mod中的module, 默认为文件夹名称)：github.com/myorg/my-microservice
请输入框架源码路径(用于 go.mod replace，默认当前目录)：D:\project\gorp

created: my-microservice
next: cd my-microservice
      go mod tidy
      go run ./cmd/app
```

## multi-flat-wire 模板目录结构（含 gRPC）

```
my-microservice/
├── proto/                   # Proto 定义（共享）
│   └── user/
│       └── v1/
│           ├── user.pb.go          # 生成的消息类型
│           └── user_grpc.pb.go     # 生成的 gRPC 代码
├── services/
│   ├── user/                # 用户服务（提供 gRPC）
│   │   ├── cmd/
│   │   │   ├── main.go              # 服务入口
│   │   │   ├── wire.go              # Wire 定义
│   │   │   └── wire_gen.go          # Wire 生成
│   │   ├── config/
│   │   │   ├── app.yaml             # 配置文件
│   │   │   └── config.go            # 配置结构
│   │   ├── internal/
│   │   │   ├── biz/
│   │   │   │   └── user.go          # 业务逻辑
│   │   │   ├── data/
│   │   │   │   └── user.go          # 数据访问
│   │   │   ├── service/
│   │   │   │   └── service.go       # 服务组装
│   │   │   └── server/
│   │   │       └── http.go          # HTTP 处理
│   │   └── start.go                 # 启动逻辑
│   ├── order/               # 订单服务（调用 user gRPC）
│   │   ├── cmd/
│   │   │   ├── main.go
│   │   │   ├── wire.go
│   │   │   └── wire_gen.go
│   │   ├── internal/
│   │   │   ├── biz/
│   │   │   │   └── order.go
│   │   │   ├── data/
│   │   │   │   └── order.go
│   │   │   └── service/
│   │   └── config/
│   │   └── start.go
│   └── product/             # 产品服务
├── go.mod                   # 共享模块定义
├── Makefile                 # 构建/Proto 生成命令
├── README.md
└── .gitignore
```

---

# 第二部分：框架 gRPC 架构

## 新架构：Proto-first gRPC

框架已升级为 **Proto-first** 主线，提供两个核心接口：

### 核心接口定义

```go
// contract/grpc.go

// GRPCConnFactory - 客户端连接工厂
// 按服务名获取框架托管的 gRPC 连接
type GRPCConnFactory interface {
    Conn(ctx context.Context, service string) (*grpc.ClientConn, error)
}

// GRPCServerRegistrar - 服务端注册器
// 注册标准 protobuf service 到框架托管的 grpc.Server
type GRPCServerRegistrar interface {
    RegisterProto(func(server *grpc.Server) error) error
    Server() *grpc.Server  // 底层 server 访问（逃生舱）
}
```

### 框架托管的能力

| 能力 | 说明 | 实现位置 |
|-----|------|---------|
| 服务发现 | 自动通过 Registry 发现服务实例 | `provider.go` Client |
| 负载均衡 | 通过 Selector 选择实例 | `provider.go` Client |
| Metadata 传播 | 自动注入/提取 metadata | 拦截器 |
| 链路追踪 | 自动注入 TraceID | 拦截器 |
| 服务认证 | 自动生成/验证 Token | 拦截器 |
| 熔断保护 | 自动熔断失败服务 | 拦截器 |
| Prometheus 指标 | 自动收集请求指标 | 拦截器 |
| 连接池 | 按地址缓存 ClientConn | Client.connPool |
| Health Check | 内置 gRPC 健康检查 | Server |
| Reflection | 内置 gRPC 反射服务 | Server |

---

# 第三部分：服务端实现（Proto-first）

## 服务端注册方式

框架模板中的标准模式：

```go
// services/user/cmd/main.go
package main

import (
    "fmt"
    "os"

    userv1 "github.com/myorg/my-microservice/proto/user/v1"
    userserver "github.com/myorg/my-microservice/services/user/internal/server"
    userdata "github.com/myorg/my-microservice/services/user/internal/data"
    frameworkbootstrap "github.com/ngq/gorp/framework/bootstrap"
    frameworkcontainer "github.com/ngq/gorp/framework/container"
    "google.golang.org/grpc"
)

func main() {
    if err := frameworkbootstrap.BootHTTPService("user-service",
        frameworkbootstrap.HTTPServiceOptions{},
        migrate, setup); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

func migrate(rt *frameworkbootstrap.HTTPServiceRuntime) error {
    return frameworkbootstrap.AutoMigrateModels(rt, &userdata.UserPO{})
}

func setup(rt *frameworkbootstrap.HTTPServiceRuntime) error {
    if rt.DB == nil {
        return fmt.Errorf("user-service requires gorm database")
    }

    // Wire 组装服务
    services, err := wireUserServices(rt.DB)
    if err != nil {
        return err
    }

    // 注册 HTTP 路由
    httpServer := userserver.NewHTTPServer(services)
    httpServer.RegisterRoutes(rt.Engine)

    // 【关键】注册 gRPC 服务（Proto-first 方式）
    registrar, err := frameworkcontainer.MakeGRPCServerRegistrar(rt.Container)
    if err != nil {
        return err
    }
    return registrar.RegisterProto(func(server *grpc.Server) error {
        userv1.RegisterUserServiceServer(server, services.User)
        return nil
    })
}
```

## 服务实现

```go
// proto/user/v1/user_grpc.pb.go（自动生成）
type UserServiceServer interface {
    GetUser(context.Context, *GetUserRequest) (*GetUserResponse, error)
    mustEmbedUnimplementedUserServiceServer()
}

// services/user/internal/service/user_service.go
package service

import (
    "context"

    userv1 "github.com/myorg/my-microservice/proto/user/v1"
    "github.com/myorg/my-microservice/services/user/internal/biz"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type UserService struct {
    userv1.UnimplementedUserServiceServer

    uc *biz.UserUseCase
}

func NewUserService(uc *biz.UserUseCase) *UserService {
    return &UserService{uc: uc}
}

func (s *UserService) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
    user, err := s.uc.GetByID(ctx, req.Id)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "用户不存在: %d", req.Id)
    }

    return &userv1.GetUserResponse{
        Id:    user.ID,
        Name:  user.Name,
        Email: user.Email,
    }, nil
}
```

---

# 第四部分：客户端调用（Proto-first）

## 获取连接并调用

```go
// services/order/cmd/main.go
package main

import (
    frameworkcontainer "github.com/ngq/gorp/framework/container"
)

func setup(rt *frameworkbootstrap.HTTPServiceRuntime) error {
    // 【关键】获取 gRPC 连接工厂
    connFactory, err := frameworkcontainer.MakeGRPCConnFactory(rt.Container)
    if err != nil {
        return err
    }

    // Wire 组装服务（传入连接工厂）
    services, err := wireOrderServices(rt.DB, connFactory)
    if err != nil {
        return err
    }

    // 注册 HTTP 路由
    httpServer := orderserver.NewHTTPServer(services)
    httpServer.RegisterRoutes(rt.Engine)
    return nil
}
```

## 使用连接调用其他服务

```go
// services/order/internal/biz/order.go
package biz

import (
    "context"

    userv1 "github.com/myorg/my-microservice/proto/user/v1"
    "github.com/ngq/gorp/framework/contract"
)

type OrderUseCase struct {
    repo         OrderRepository
    userConnFactory contract.GRPCConnFactory  // 注入连接工厂
}

func NewOrderUseCase(repo OrderRepository, userConnFactory contract.GRPCConnFactory) *OrderUseCase {
    return &OrderUseCase{
        repo:             repo,
        userConnFactory:  userConnFactory,
    }
}

// CreateOrder 创建订单（调用 user 服务验证用户）
func (uc *OrderUseCase) CreateOrder(ctx context.Context, userID uint, productID uint) (*Order, error) {
    // 【关键】获取 user 服务的连接
    conn, err := uc.userConnFactory.Conn(ctx, "user-service")
    if err != nil {
        return nil, err
    }

    // 【关键】使用 Proto 生成的客户端调用（类型安全）
    userClient := userv1.NewUserServiceClient(conn)
    user, err := userClient.GetUser(ctx, &userv1.GetUserRequest{Id: uint64(userID)})
    if err != nil {
        return nil, err
    }

    // 业务逻辑：验证用户后创建订单
    order := &Order{
        UserID:    userID,
        ProductID: productID,
        // 使用返回的用户信息
        UserName:  user.Name,
        Status:    "pending",
    }

    err = uc.repo.Create(ctx, order)
    return order, err
}
```

## Wire 依赖注入配置

```go
// services/order/cmd/wire.go
//go:build wireinject

package main

import (
    "github.com/ngq/gorp/framework/contract"
    "github.com/myorg/my-microservice/services/order/internal/service"
    "github.com/google/wire"
    "gorm.io/gorm"
)

func wireOrderServices(
    db *gorm.DB,
    userConnFactory contract.GRPCConnFactory,  // 注入连接工厂
    dlock contract.DistributedLock,
) (*service.Services, error) {
    panic(wire.Build(
        // 数据层 ProviderSet
        orderdata.ProviderSet,
        // 业务层 ProviderSet（依赖 userConnFactory）
        orderbiz.ProviderSet,
        // 服务层 ProviderSet
        orderservice.ProviderSet,
    ))
}
```

---

# 第五部分：配置详解

## gRPC 配置

```yaml
# services/user/config/app.yaml
app:
  name: user-service
  env: development
  address: :8080

# gRPC 配置
rpc:
  mode: grpc
  timeout_ms: 30000
  grpc:
    # 服务端监听地址
    address: :9090
    # 是否使用 insecure（开发环境）
    insecure: true

# 服务发现配置（生产环境）
registry:
  type: consul
  address: "127.0.0.1:8500"

# 数据库配置
database:
  driver: sqlite
  dsn: user.db

# 日志配置
log:
  level: debug
  format: console
```

## 配置项说明

| 配置项 | 类型 | 说明 | 默认值 |
|-------|------|------|-------|
| `rpc.mode` | string | RPC 模式：grpc/http/noop | grpc |
| `rpc.timeout_ms` | int | 调用超时（毫秒） | 30000 |
| `rpc.grpc.address` | string | gRPC 服务端监听地址 | :9090 |
| `rpc.grpc.insecure` | bool | 是否跳过 TLS | true |
| `registry.type` | string | 注册中心类型 | noop |
| `registry.address` | string | 注册中心地址 | - |

---

# 第六部分：Proto 文件管理

## Proto 定义

```protobuf
// proto/user/v1/user.proto
syntax = "proto3";

package user.v1;

option go_package = "github.com/myorg/my-microservice/proto/user/v1;v1";

service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse);
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
    rpc ListUsers(ListUsersRequest) returns (ListUsersResponse);
}

message GetUserRequest {
    uint64 id = 1;
}

message GetUserResponse {
    uint64 id = 1;
    string name = 2;
    string email = 3;
}

message CreateUserRequest {
    string name = 1;
    string email = 2;
}

message CreateUserResponse {
    uint64 id = 1;
}

message ListUsersRequest {
    int32 page = 1;
    int32 page_size = 2;
}

message ListUsersResponse {
    repeated GetUserResponse users = 1;
    int32 total = 2;
}
```

## Makefile 生成命令

```makefile
# Makefile（模板自带）
.PHONY: proto proto-gen proto-clean

# 生成 Proto 代码
proto:
	protoc --go_out=. --go-grpc_out=. \
		-I=proto \
		proto/user/v1/*.proto \
		proto/order/v1/*.proto

# 生成并格式化
proto-gen: proto
	go fmt ./proto/...

# 清理
proto-clean:
	rm -rf proto/**/*.pb.go proto/**/*_grpc.pb.go

# Wire 生成
wire:
	cd services/user/cmd && wire
	cd services/order/cmd && wire

# 完整构建
build: proto-gen wire
	go build ./...
```

## 使用 gorp proto 命令

```bash
# 查看帮助
gorp proto --help

# 生成 Proto 代码
gorp proto gen

# 指定输入/输出目录
gorp proto gen --input=proto --output=.
```

---

# 第七部分：拦截器与中间件

## 框架自动注入的拦截器

框架在创建 gRPC Server/Client 时自动注入以下拦截器：

### 服务端拦截器

```go
// framework/provider/rpc/grpc/provider.go - Server.newGRPCServer()

// 1. 指标收集拦截器（内置）
appgrpc.UnaryServerInterceptor()  // Prometheus 指标 + TraceID 提取

// 2. 链路追踪拦截器（可选）
tracingmw.UnaryServerInterceptor(tracer, serviceName)

// 3. Metadata 传播拦截器（可选）
metadatamw.UnaryServerInterceptor(propagator)

// 4. 服务认证拦截器（可选）
serviceauthtoken.UnaryServerInterceptor(authenticator)
```

### 客户端拦截器

```go
// framework/provider/rpc/grpc/provider.go - Client.getConn()

// 1. Metadata 传播拦截器
metadatamw.UnaryClientInterceptor(propagator)

// 2. 链路追踪拦截器
tracingmw.UnaryClientInterceptor(tracer, serviceName)

// 3. 服务认证拦截器
serviceauthtoken.UnaryClientInterceptor(serviceAuth, serviceName)

// 4. 熔断拦截器
circuitBreakerUnaryInterceptor(serviceName)

// 5. 指标收集拦截器（内置）
appgrpc.UnaryClientInterceptor()  // TraceID 注入
```

## Prometheus 指标

框架自动收集的 gRPC 指标：

```
# 请求总数
gorp_grpc_requests_total{method="/user.v1.UserService/GetUser",status="OK"}

# 请求耗时分布
gorp_grpc_request_duration_seconds{method="/user.v1.UserService/GetUser",status="OK"}

# 当前处理中的请求数
gorp_grpc_requests_in_flight{method="/user.v1.UserService/GetUser"}
```

## 获取 TraceID

```go
import "github.com/ngq/gorp/app/grpc"

func (s *UserService) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
    // 获取 TraceID
    traceID := grpc.GetTraceID(ctx)

    // 获取 RequestID
    requestID := grpc.GetRequestID(ctx)

    s.logger.Info(ctx, "GetUser called", map[string]any{
        "trace_id":   traceID,
        "request_id": requestID,
    })

    // 业务逻辑...
}
```

---

# 第八部分：对比与选择

## 新架构 vs 旧架构对比

| 特性 | 新架构（Proto-first） | 旧架构（RPCClient.Call） |
|-----|----------------------|------------------------|
| 服务发现 | ✅ 自动托管 | ✅ 自动托管 |
| 负载均衡 | ✅ 自动托管 | ✅ 自动托管 |
| 链路追踪 | ✅ 自动注入 | ✅ 自动注入 |
| 熔断保护 | ✅ 自动熔断 | ❌ 无 |
| 类型安全 | ✅ 强类型（Proto 生成） | ❌ 弱类型（interface{}） |
| IDE 支持 | ✅ 完整提示 | ❌ 无提示 |
| 编译检查 | ✅ 编译时检查 | ❌ 运行时发现 |
| 方法名 | ✅ 自动生成 | ❌ 手动拼接 |
| Health Check | ✅ 内置 | ❌ 无 |
| Reflection | ✅ 内置 | ❌ 无 |

## 开发难度对比

| 方案 | 初次学习 | 日常开发 | 维护成本 | 推荐度 |
|-----|---------|---------|---------|-------|
| Proto-first | ⭐⭐⭐ 中 | ⭐ 低 | ⭐ 低 | **强烈推荐** |
| RPCClient.Call | ⭐⭐ 低 | ⭐⭐⭐⭐ 高 | ⭐⭐⭐⭐ 高 | 不推荐 |
| 标准 gRPC（无框架） | ⭐⭐⭐⭐ 高 | ⭐⭐ 中 | ⭐⭐⭐ 中 | 外部对接 |

---

# 第九部分：快速启动指南

## 完整流程

```bash
# 1. 安装 gorp 命令
go install ./cmd/gorp

# 2. 创建多服务项目（含 gRPC 示例）
gorp new multi-wire
# 输入：my-microservice
# 输入：github.com/myorg/my-microservice
# 输入：D:\project\gorp（框架路径）

# 3. 进入项目目录
cd my-microservice

# 4. 下载依赖
go mod tidy

# 5. 生成 Wire 代码
make wire
# 或：cd services/user/cmd && wire

# 6. 启动 user 服务
go run ./services/user/cmd

# 7. 启动 order 服务（调用 user）
go run ./services/order/cmd

# 8. 测试 gRPC
grpcurl -plaintext -d '{"id": 1}' localhost:9090 user.v1.UserService/GetUser
```

## 验证服务

```bash
# 查看 gRPC 服务列表
grpcurl -plaintext localhost:9090 list

# 查看服务描述
grpcurl -plaintext localhost:9090 describe user.v1.UserService

# 调用方法
grpcurl -plaintext -d '{"id": 123}' localhost:9090 user.v1.UserService/GetUser

# 健康检查
grpcurl -plaintext localhost:9090 grpc.health.v1.Health/Check
```

---

# 第十部分：常见问题

## Q1: gRPC Server 端口是什么？

默认端口 `:9090`，由配置 `rpc.grpc.address` 控制。HTTP 服务端口 `:8080`，两者独立。

## Q2: 如何添加新的 Proto 服务？

```bash
# 1. 创建 Proto 文件
mkdir -p proto/product/v1
vim proto/product/v1/product.proto

# 2. 生成代码
make proto

# 3. 实现服务
# services/product/internal/service/product_service.go

# 4. 注册服务
registrar.RegisterProto(func(server *grpc.Server) error {
    productv1.RegisterProductServiceServer(server, services.Product)
    return nil
})
```

## Q3: 客户端如何调用多个服务？

```go
// 注入连接工厂后，可以获取任意服务的连接
conn, err := connFactory.Conn(ctx, "user-service")   // 用户服务
conn, err := connFactory.Conn(ctx, "order-service")  // 订单服务
conn, err := connFactory.Conn(ctx, "product-service") // 产品服务
```

## Q4: 如何处理连接错误？

框架会自动重试和熔断。业务层只需处理调用错误：

```go
conn, err := connFactory.Conn(ctx, "user-service")
if err != nil {
    // 服务不可用（可能被熔断或无实例）
    return nil, fmt.Errorf("服务暂不可用: %w", err)
}

resp, err := userClient.GetUser(ctx, req)
if err != nil {
    // 业务错误
    return nil, err
}
```

## Q5: 如何调试 gRPC？

```bash
# 使用 grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 查看服务
grpcurl -plaintext localhost:9090 list

# 调用方法
grpcurl -plaintext -d '{"id": 1}' localhost:9090 user.v1.UserService/GetUser

# 使用 grpcui（Web UI）
go install github.com/fullstorydev/grpcui/cmd/grpcui@latest
grpcui -plaintext localhost:9090
```

---

# 总结

| 场景 | 推荐方案 | 命令 |
|-----|---------|------|
| 微服务项目 | Proto-first + multi-flat-wire | `gorp new multi-wire` |
| 单体项目 | golayout | `gorp new` |
| 需要依赖注入 | Wire 模板 | `gorp new wire` |
| 外部系统对接 | 标准 gRPC | 无需框架托管 |
| 快速原型 | golayout + 封装层 | `gorp new` |

**最佳实践**：使用 `gorp new multi-wire` 创建项目，框架自动托管服务发现、熔断、追踪，业务代码保持标准 Proto-first 模式。