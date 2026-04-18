# gRPC 开发指南

本文档详细说明在框架中使用 gRPC 的两种方式及其选择建议，基于框架真实 API 进行说明。

## 目录

- [概述](#概述)
- [框架架构分析](#框架架构分析)
- [方式对比](#方式对比)
- [框架内置 gRPC Provider](#框架内置-grpc-provider)
- [外部 Proto 标准 gRPC](#外部-proto-标准-grpc)
- [混合方案（推荐）](#混合方案推荐)
- [完整项目示例](#完整项目示例)
- [配置详解](#配置详解)
- [拦截器与中间件](#拦截器与中间件)
- [服务发现集成](#服务发现集成)
- [链路追踪集成](#链路追踪集成)
- [最佳实践](#最佳实践)
- [常见问题](#常见问题)

---

## 概述

框架提供了两种 gRPC 使用方式：

1. **框架内置 gRPC Provider** - 封装的统一 RPC 接口（`contract.RPCClient`）
2. **外部 Proto 标准 gRPC** - 使用 protobuf 生成的标准客户端

两种方式各有优劣，选择取决于具体使用场景。

---

## 框架架构分析

### 核心接口定义

框架在 `contract/rpc.go` 中定义了 RPC 抽象：

```go
// RPCClient 定义 RPC 客户端抽象
type RPCClient interface {
    // Call 执行 RPC 调用
    // - service: 目标服务名称（如 "user-service"）
    // - method: 方法名称（HTTP 模式为 path，gRPC 模式为方法名）
    // - req: 请求对象
    // - resp: 响应对象（指针）
    Call(ctx context.Context, service, method string, req, resp any) error

    // CallRaw 执行原始数据 RPC 调用
    CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error)

    // Close 关闭客户端连接
    Close() error
}

// RPCServer 定义 RPC 服务端抽象
type RPCServer interface {
    // Register 注册服务处理器
    Register(service string, handler any) error

    // Start 启动 RPC 服务
    Start(ctx context.Context) error

    // Stop 停止 RPC 服务
    Stop(ctx context.Context) error

    // Addr 返回服务监听地址
    Addr() string
}
```

### Provider 实现层级

框架提供三种 RPC Provider 实现：

```
framework/provider/rpc/
├── noop/    # 单体项目空实现（零依赖）
├── http/    # HTTP REST API 实现
└── grpc/    # gRPC/Protobuf 实现
```

切换方式只需修改配置：

```yaml
# config.yaml
rpc:
  mode: grpc  # 可选: noop, http, grpc
```

---

## 方式对比

### 功能对比表

| 特性 | 框架内置 Provider | 外部 Proto gRPC |
|-----|------------------|-----------------|
| 服务发现 | ✅ 自动集成（Registry + Selector） | ❌ 需自行实现 |
| 负载均衡 | ✅ 自动集成 | ❌ 需自行实现 |
| 链路追踪 | ✅ 自动注入 TraceID | ❌ 需自行集成 |
| 服务认证 | ✅ 自动注入 Token | ❌ 需自行处理 |
| Metadata 透传 | ✅ 自动传播 | ❌ 需手动处理 |
| 连接池管理 | ✅ 自动缓存连接 | ❌ 需手动管理 |
| Prometheus 指标 | ✅ 自动收集 | ❌ 需自行集成 |
| 类型安全 | ❌ 弱类型（interface{}） | ✅ 强类型 |
| IDE 支持 | ❌ 无代码提示 | ✅ 完整代码提示 |
| 编译检查 | ❌ 运行时发现错误 | ✅ 编译时检查 |
| 开发效率 | ⭐⭐ 较低 | ⭐⭐⭐⭐⭐ 高 |

### 适用场景

| 场景 | 推荐方案 |
|-----|---------|
| 微服务内部高频调用 | 框架 Provider（功能集成完整） |
| 需要类型安全 + IDE 支持 | 标准 Proto + 框架连接池 |
| 与外部系统对接 | 标准 Proto gRPC |
| 快速原型开发 | 框架 Provider + 封装层 |
| 单体项目 | noop Provider（无服务间调用） |

---

## 框架内置 gRPC Provider

### 架构说明

框架 gRPC Provider 位于 `framework/provider/rpc/grpc/provider.go`，核心特性：

1. **自动服务发现** - 通过 `ServiceRegistry` 发现服务实例
2. **负载均衡** - 通过 `Selector` 选择实例
3. **Metadata 传播** - 自动注入/提取 metadata
4. **链路追踪** - 自动注入 TraceID
5. **服务认证** - 自动生成/验证 Token
6. **连接池** - 按地址缓存 `grpc.ClientConn`

### 客户端核心实现

```go
// Client 是 gRPC RPC 客户端实现
type Client struct {
    cfg                *contract.RPCConfig
    registry           contract.ServiceRegistry    // 服务发现
    selector           contract.Selector           // 负载均衡
    metadataPropagator contract.MetadataPropagator // Metadata 传播
    serviceAuth        contract.ServiceAuthenticator // 服务认证
    tracer             contract.Tracer             // 链路追踪

    connPool sync.Map // 连接池: map[string]*grpc.ClientConn
}

// Call 执行 RPC 调用
func (c *Client) Call(ctx context.Context, service, method string, req, resp any) error {
    // 1. 获取连接（支持服务发现）
    conn, done, err := c.getConn(ctx, service)
    if err != nil {
        return fmt.Errorf("rpc: get connection failed: %w", err)
    }

    // 2. 设置超时
    timeout := time.Duration(c.cfg.TimeoutMS) * time.Millisecond
    if timeout > 0 {
        ctx, cancel := context.WithTimeout(ctx, timeout)
        defer cancel()
    }

    // 3. 准备 outgoing metadata
    md := metadata.New(nil)

    // 4. 注入 metadata、tracing、认证信息
    if c.metadataPropagator != nil {
        c.metadataPropagator.Inject(ctx, carrier)
    }
    if c.tracer != nil {
        c.tracer.Inject(ctx, carrier)
    }
    if c.serviceAuth != nil {
        token, _ := c.serviceAuth.GenerateToken(ctx, service)
        md.Set("x-service-token", token)
    }

    // 5. 发起 gRPC 调用
    ctx = metadata.NewOutgoingContext(ctx, md)
    return conn.Invoke(ctx, method, req, resp)
}

// getConn 获取或创建服务连接（支持服务发现）
func (c *Client) getConn(ctx context.Context, service string) (*grpc.ClientConn, contract.DoneFunc, error) {
    // 1. 通过 Registry 发现服务实例
    addr, done, err := c.resolveTarget(ctx, service)

    // 2. 从连接池缓存获取
    if cached, ok := c.connPool.Load(addr); ok {
        conn := cached.(*grpc.ClientConn)
        if conn.GetState().String() != "SHUTDOWN" {
            return conn, done, nil
        }
    }

    // 3. 创建新连接
    conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    c.connPool.Store(addr, conn)
    return conn, done, nil
}
```

### 服务端核心实现

```go
// Server 是 gRPC RPC 服务端实现
type Server struct {
    server   *grpc.Server
    listener net.Listener

    // 自动集成拦截器
    unaryInterceptors []grpc.UnaryServerInterceptor
}

// Start 启动 gRPC 服务
func (s *Server) Start(ctx context.Context) error {
    // 1. 创建监听器
    lis, err := net.Listen("tcp", s.cfg.Address)

    // 2. 创建 gRPC Server（自动注入拦截器）
    opts := []grpc.ServerOption{}

    // 添加链路追踪拦截器
    if tracer != nil {
        unaryInterceptors = append(unaryInterceptors, tracingmw.UnaryServerInterceptor(tracer))
    }

    // 添加 Metadata 传播拦截器
    if propagator != nil {
        unaryInterceptors = append(unaryInterceptors, metadatamw.UnaryServerInterceptor(propagator))
    }

    // 添加服务认证拦截器
    if authenticator != nil {
        unaryInterceptors = append(unaryInterceptors, serviceauthtoken.UnaryServerInterceptor(authenticator))
    }

    opts = append(opts, grpc.ChainUnaryInterceptor(unaryInterceptors...))
    s.server = grpc.NewServer(opts...)

    // 3. 启动服务（非阻塞）
    go s.server.Serve(lis)
    return nil
}

// GRPCServer 返回底层 grpc.Server（用于注册 protobuf service）
func (s *Server) GRPCServer() *grpc.Server {
    return s.server
}
```

### 使用示例

#### 1. 基本调用

```go
package main

import (
    "context"
    "github.com/ngq/gorp/framework/container"
    "github.com/ngq/gorp/framework/contract"
)

func main() {
    // 从容器获取 RPC 客户端
    rpcClient := container.MustMake[contract.RPCClient](c, contract.RPCClientKey)

    // 发起调用
    req := &GetUserReq{Id: 123}
    var resp GetUserResp

    err := rpcClient.Call(
        context.Background(),
        "user-service",      // 服务名
        "/api.user.v1.UserService/GetUser", // gRPC 方法全名
        req,
        &resp,
    )
}
```

#### 2. 项目中的封装示例

项目中已有的客户端封装（`examples/nop-go/shared/rpc/clients.go`）：

```go
// InventoryClient 库存服务客户端
type InventoryClient struct {
    client contract.RPCClient
}

// NewInventoryClient 创建库存服务客户端
func NewInventoryClient(client contract.RPCClient) *InventoryClient {
    return &InventoryClient{client: client}
}

// ReserveStock 预留库存
func (c *InventoryClient) ReserveStock(ctx context.Context, req *inventory.ReserveStockRequest) (*inventory.ReserveStockResponse, error) {
    resp := &inventory.ReserveStockResponse{}
    err := c.client.Call(ctx, ServiceInventory, "ReserveStock", req, resp)
    if err != nil {
        resp.Success = false
        resp.ErrorMessage = err.Error()
    }
    return resp, err
}
```

### 方法名规则

| 模式 | service 参数 | method 参数 | 示例 |
|-----|-------------|------------|------|
| HTTP | 服务名（映射 baseURL） | HTTP path | `Call(ctx, "user-service", "/api/user/get", req, resp)` |
| gRPC | 服务名（用于服务发现） | gRPC 方法全名 | `Call(ctx, "user-service", "/user.UserService/GetUser", req, resp)` |

**gRPC 方法全名格式**：`/{package}.{ServiceName}/{MethodName}`

例如：
- `/api.user.v1.UserService/GetUser`
- `/inventory.InventoryService/ReserveStock`

### 优点

1. **开箱即用的微服务能力**
   - 自动服务注册与发现
   - 内置负载均衡策略（通过 Selector）
   - 集成链路追踪（Tracer）
   - 统一的服务认证（ServiceAuthenticator）
   - Metadata 自动传播

2. **连接池管理**
   - 按地址缓存 `grpc.ClientConn`
   - 自动复用连接
   - 健康状态检查
   - 自动关闭无效连接

3. **统一接口**
   - HTTP/gRPC 切换只需改配置
   - 无需修改业务代码

### 缺点

1. **需要手动拼接方法名**
   ```go
   // ❌ 方法名容易写错，运行时才发现
   err := client.Call(ctx, "user-service", "/api.user.v1.UserService/GetUserr", req, resp)
   // 写错任何字符都会导致 Unimplemented 错误
   ```

2. **弱类型，缺少编译检查**
   ```go
   var resp GetUserResp
   err := client.Call(ctx, "user-service", "/xxx", req, &resp)
   // req 和 resp 类型不匹配，编译器无法发现
   // 运行时才会报序列化错误
   ```

3. **无 IDE 代码提示**
   - 方法名是字符串，无自动补全
   - 字段名需要查文档或 proto 文件
   - 无法跳转到定义

---

## 外部 Proto 标准 gRPC

### 定义 Proto 文件

```protobuf
// api/user/v1/user.proto
syntax = "proto3";

package api.user.v1;

option go_package = "github.com/your-org/project/api/user/v1;v1";

service UserService {
    rpc GetUser(GetUserReq) returns (GetUserResp);
    rpc CreateUser(CreateUserReq) returns (CreateUserResp);
    rpc ListUsers(ListUsersReq) returns (ListUsersResp);
    rpc UpdateUser(UpdateUserReq) returns (UpdateUserResp);
    rpc DeleteUser(DeleteUserReq) returns (DeleteUserResp);
}

message GetUserReq {
    int64 id = 1;
}

message GetUserResp {
    int64 id = 1;
    string name = 2;
    string email = 3;
    string phone = 4;
    int32 status = 5;
    int64 created_at = 6;
    int64 updated_at = 7;
}

message CreateUserReq {
    string name = 1;
    string email = 2;
    string phone = 3;
}

message CreateUserResp {
    int64 id = 1;
}

message ListUsersReq {
    int32 page = 1;
    int32 page_size = 2;
    string keyword = 3;
}

message ListUsersResp {
    repeated GetUserResp users = 1;
    int32 total = 2;
    int32 page = 3;
    int32 page_size = 4;
}

message UpdateUserReq {
    int64 id = 1;
    string name = 2;
    string email = 3;
    string phone = 4;
}

message UpdateUserResp {
    bool success = 1;
}

message DeleteUserReq {
    int64 id = 1;
}

message DeleteUserResp {
    bool success = 1;
}
```

### 生成代码

```bash
# 安装 protoc 和插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成代码
protoc --go_out=. --go-grpc_out=. \
    -I=api \
    api/user/v1/user.proto

# 生成的文件：
# api/user/v1/user.pb.go        - 消息类型
# api/user/v1/user_grpc.pb.go   - gRPC 服务定义
```

### 生成的代码结构

```go
// user_grpc.pb.go 自动生成的客户端接口
type UserServiceClient interface {
    GetUser(ctx context.Context, in *GetUserReq, opts ...grpc.CallOption) (*GetUserResp, error)
    CreateUser(ctx context.Context, in *CreateUserReq, opts ...grpc.CallOption) (*CreateUserResp, error)
    ListUsers(ctx context.Context, in *ListUsersReq, opts ...grpc.CallOption) (*ListUsersResp, error)
    UpdateUser(ctx context.Context, in *UpdateUserReq, opts ...grpc.CallOption) (*UpdateUserResp, error)
    DeleteUser(ctx context.Context, in *DeleteUserReq, opts ...grpc.CallOption) (*DeleteUserResp, error)
}

// 服务端接口（需要实现）
type UserServiceServer interface {
    GetUser(context.Context, *GetUserReq) (*GetUserResp, error)
    CreateUser(context.Context, *CreateUserReq) (*CreateUserResp, error)
    ListUsers(context.Context, *ListUsersReq) (*ListUsersResp, error)
    UpdateUser(context.Context, *UpdateUserReq) (*UpdateUserResp, error)
    DeleteUser(context.Context, *DeleteUserReq) (*DeleteUserResp, error)
    mustEmbedUnimplementedUserServiceServer()
}
```

### 服务端实现

参考项目中的实际实现（`examples/nop-go/services/customer-service/internal/grpc/customer_grpc.go`）：

```go
package grpcsvc

import (
    "context"
    pb "your-project/api/user/v1"
)

// UserServiceServer 用户服务 gRPC 服务端
type UserServiceServer struct {
    pb.UnimplementedUserServiceServer

    // 注入依赖
    userRepo repository.UserRepository
    logger   contract.Logger
}

// NewUserServiceServer 创建用户 gRPC 服务端
func NewUserServiceServer(userRepo repository.UserRepository, logger contract.Logger) *UserServiceServer {
    return &UserServiceServer{
        userRepo: userRepo,
        logger:   logger,
    }
}

// GetUser 获取用户
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserReq) (*pb.GetUserResp, error) {
    user, err := s.userRepo.FindByID(ctx, req.Id)
    if err != nil {
        s.logger.Error(ctx, "获取用户失败", map[string]any{
            "id":  req.Id,
            "err": err.Error(),
        })
        return nil, status.Errorf(codes.NotFound, "用户不存在: %d", req.Id)
    }

    return &pb.GetUserResp{
        Id:        user.ID,
        Name:      user.Name,
        Email:     user.Email,
        Phone:     user.Phone,
        Status:    int32(user.Status),
        CreatedAt: user.CreatedAt.Unix(),
        UpdatedAt: user.UpdatedAt.Unix(),
    }, nil
}

// CreateUser 创建用户
func (s *UserServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserResp, error) {
    user := &model.User{
        Name:  req.Name,
        Email: req.Email,
        Phone: req.Phone,
    }

    err := s.userRepo.Create(ctx, user)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "创建用户失败: %v", err)
    }

    return &pb.CreateUserResp{Id: user.ID}, nil
}

// ListUsers 列出用户
func (s *UserServiceServer) ListUsers(ctx context.Context, req *pb.ListUsersReq) (*pb.ListUsersResp, error) {
    users, total, err := s.userRepo.List(ctx, req.Page, req.PageSize, req.Keyword)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "查询用户失败: %v", err)
    }

    resp := &pb.ListUsersResp{
        Total:    int32(total),
        Page:     req.Page,
        PageSize: req.PageSize,
    }

    for _, user := range users {
        resp.Users = append(resp.Users, &pb.GetUserResp{
            Id:        user.ID,
            Name:      user.Name,
            Email:     user.Email,
        })
    }

    return resp, nil
}
```

### 标准客户端调用

```go
package main

import (
    "context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "your-project/api/user/v1"
)

func main() {
    // 建立连接
    conn, err := grpc.Dial("localhost:9090",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithUnaryInterceptor(UnaryClientInterceptor()), // 可选：添加拦截器
    )
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // 创建客户端
    client := pb.NewUserServiceClient(conn)

    // 调用方法 - 类型安全，IDE 有提示
    resp, err := client.GetUser(context.Background(), &pb.GetUserReq{Id: 123})
    if err != nil {
        panic(err)
    }

    fmt.Printf("User: %s\n", resp.Name)
}
```

### 优点

1. **类型安全**
   ```go
   // ✅ 编译器检查参数类型
   resp, err := client.GetUser(ctx, &pb.GetUserReq{Id: 123})

   // ❌ 编译错误：字段名拼写错误
   resp, err := client.GetUser(ctx, &pb.GetUserReq{Idxxx: 123})

   // ❌ 编译错误：方法不存在
   resp, err := client.GetUserr(ctx, req)
   ```

2. **完整的 IDE 支持**
   - 方法名自动补全
   - 参数字段提示
   - 跳转到定义
   - 重构支持（重命名方法/字段）

3. **标准化工具链**
   - protoc 代码生成
   - gRPC 官方生态
   - 跨语言支持（Java/Python/Node.js 等）

### 缺点

1. **需要额外管理连接**
   ```go
   // 每个服务都需要手动创建连接
   userConn, _ := grpc.Dial("user-service:9090", ...)
   orderConn, _ := grpc.Dial("order-service:9090", ...)
   defer userConn.Close()
   defer orderConn.Close()
   ```

2. **缺少服务发现集成**
   - 硬编码服务地址
   - 需要自行实现服务发现
   - 负载均衡需手动配置

3. **缺少框架集成**
   - 无链路追踪注入
   - 无 Metadata 传播
   - 无服务认证

---

## 混合方案（推荐）

### 方案概述

结合框架的服务发现、连接池管理与 Proto 生成的类型安全客户端，获得两者的优点。

**核心思路**：使用框架获取 gRPC 连接，使用 Proto 生成的客户端进行类型安全调用。

### 实现方案

#### 方案一：扩展框架 RPCClient

在框架 `contract.RPCClient` 基础上添加获取底层连接的方法：

```go
// contract/rpc.go 扩展
type RPCClient interface {
    Call(ctx context.Context, service, method string, req, resp any) error
    CallRaw(ctx context.Context, service, method string, data []byte) ([]byte, error)
    Close() error

    // 新增：获取底层 gRPC 连接
    GRPCConn(ctx context.Context, service string) (*grpc.ClientConn, error)
}

// framework/provider/rpc/grpc/provider.go 实现
func (c *Client) GRPCConn(ctx context.Context, service string) (*grpc.ClientConn, error) {
    conn, _, err := c.getConn(ctx, service)
    return conn, err
}
```

使用示例：

```go
// 获取框架管理的连接
rpcClient := container.MustMake[contract.RPCClient](c, contract.RPCClientKey)

conn, err := rpcClient.GRPCConn(ctx, "user-service")
if err != nil {
    return err
}

// 使用 Proto 生成的客户端（类型安全）
userClient := pb.NewUserServiceClient(conn)
resp, err := userClient.GetUser(ctx, &pb.GetUserReq{Id: 123})
```

#### 方案二：创建连接管理器（无需修改框架）

```go
// internal/client/manager.go
package client

import (
    "context"
    "sync"

    "github.com/ngq/gorp/framework/contract"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

// ConnectionManager 连接管理器
// 使用框架的服务发现能力，为 Proto 客户端提供连接
type ConnectionManager struct {
    rpcClient contract.RPCClient

    // Proto 专用连接池
    protoConns sync.Map // map[string]*grpc.ClientConn
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(rpcClient contract.RPCClient) *ConnectionManager {
    return &ConnectionManager{
        rpcClient: rpcClient,
    }
}

// GetConn 获取服务连接（支持服务发现）
func (m *ConnectionManager) GetConn(ctx context.Context, serviceName string) (*grpc.ClientConn, error) {
    // 尝试从缓存获取
    if cached, ok := m.protoConns.Load(serviceName); ok {
        conn := cached.(*grpc.ClientConn)
        if conn.GetState().String() != "SHUTDOWN" {
            return conn, nil
        }
        m.protoConns.Delete(serviceName)
    }

    // 方式1：通过框架 RPC 调用获取地址
    // 使用框架的服务发现能力
    // 注意：这里需要框架提供获取地址的方法，或者自行实现

    // 方式2：直接创建连接（需要知道地址）
    // 如果有服务发现组件，可以获取地址
    addr := m.resolveAddress(ctx, serviceName)

    conn, err := grpc.NewClient(addr,
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithUnaryInterceptor(m.createClientInterceptor()),
    )
    if err != nil {
        return nil, err
    }

    m.protoConns.Store(serviceName, conn)
    return conn, nil
}

// resolveAddress 解析服务地址
// 可以复用框架的服务发现逻辑
func (m *ConnectionManager) resolveAddress(ctx context.Context, serviceName string) string {
    // 从配置获取服务发现地址
    // 或者使用框架的 Registry
    return serviceName // 默认返回服务名（依赖 DNS 或本地解析）
}

// createClientInterceptor 创建客户端拦截器
// 注入 TraceID、Metadata 等
func (m *ConnectionManager) createClientInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{},
        cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

        // 注入 TraceID
        if tid := ctx.Value("trace_id"); tid != nil {
            md := metadata.New(map[string]string{"x-trace-id": tid.(string)})
            ctx = metadata.NewOutgoingContext(ctx, md)
        }

        return invoker(ctx, method, req, reply, cc, opts...)
    }
}

// Close 关闭所有连接
func (m *ConnectionManager) Close() error {
    m.protoConns.Range(func(key, value interface{}) bool {
        if conn, ok := value.(*grpc.ClientConn); ok {
            conn.Close()
        }
        return true
    })
    return nil
}
```

#### 方案三：类型安全的客户端封装（项目中已有）

项目中已有的封装方式（`examples/nop-go/shared/rpc/clients.go`）：

```go
// InventoryClient 库存服务客户端
type InventoryClient struct {
    client contract.RPCClient
}

func NewInventoryClient(client contract.RPCClient) *InventoryClient {
    return &InventoryClient{client: client}
}

// ReserveStock 预留库存
func (c *InventoryClient) ReserveStock(ctx context.Context, req *inventory.ReserveStockRequest) (*inventory.ReserveStockResponse, error) {
    resp := &inventory.ReserveStockResponse{}
    err := c.client.Call(ctx, ServiceInventory, "ReserveStock", req, resp)

    // 错误处理
    if err != nil {
        resp.Success = false
        resp.ErrorMessage = err.Error()
    }
    return resp, err
}

// ConfirmStock 确认库存
func (c *InventoryClient) ConfirmStock(ctx context.Context, req *inventory.ConfirmStockRequest) (*inventory.ConfirmStockResponse, error) {
    resp := &inventory.ConfirmStockResponse{}
    err := c.client.Call(ctx, ServiceInventory, "ConfirmStock", req, resp)
    return resp, err
}

// ReleaseStock 释放库存
func (c *InventoryClient) ReleaseStock(ctx context.Context, req *inventory.ReleaseStockRequest) (*inventory.ReleaseStockResponse, error) {
    resp := &inventory.ReleaseStockResponse{}
    err := c.client.Call(ctx, ServiceInventory, "ReleaseStock", req, resp)
    return resp, err
}
```

**优点**：
- 封装后使用类型安全：`client.ReserveStock(ctx, req)`
- 无需修改框架
- 复用框架的服务发现、连接池

**缺点**：
- 需要手动编写封装代码
- 每个方法都需要封装

### 服务端集成

服务端同时支持框架集成和 Proto 注册：

```go
package main

import (
    "context"

    "github.com/ngq/gorp/framework/bootstrap"
    "github.com/ngq/gorp/framework/container"
    "github.com/ngq/gorp/framework/contract"
    "google.golang.org/grpc"

    pb "your-project/api/user/v1"
    grpcsvc "your-project/internal/grpc"
)

func main() {
    // 使用框架启动 HTTP 服务
    runtime := bootstrap.NewHTTPServiceRuntime(&bootstrap.HTTPServiceOptions{
        Name:    "user-service",
        Address: ":8080",
    })

    // 获取 RPC Server
    rpcServer := container.MustMake[contract.RPCServer](runtime.Container(), contract.RPCServerKey)

    // 启动 gRPC 服务
    if err := rpcServer.Start(context.Background()); err != nil {
        panic(err)
    }

    // 注册 Proto 生成的服务实现
    // 获取底层 grpc.Server
    grpcServer := rpcServer.GRPCServer() // 需要框架暴露此方法
    pb.RegisterUserServiceServer(grpcServer, grpcsvc.NewUserServiceServer())

    // 启动 HTTP 服务
    bootstrap.RunHTTP(runtime)
}
```

---

## 完整项目示例

### 目录结构

```
project/
├── api/                          # Proto 定义和生成代码
│   └── user/
│       └── v1/
│           ├── user.proto        # Proto 定义文件
│           ├── user.pb.go        # 生成的消息类型
│           └── user_grpc.pb.go   # 生成的 gRPC 代码
├── internal/
│   ├── client/                   # 客户端封装
│   │   ├── manager.go            # 连接管理器
│   │   ├── user_client.go        # 用户服务客户端
│   │   └── order_client.go       # 订单服务客户端
│   ├── grpc/                     # gRPC 服务实现
│   │   ├── user_grpc.go          # 用户 gRPC 服务
│   │   └── interceptor.go        # 服务端拦截器
│   ├── repository/               # 数据访问层
│   │   └── user_repo.go
│   └── service/                  # 业务逻辑层
│       └── user_service.go
├── cmd/
│   └── server/
│       └── main.go               # 服务启动入口
├── config/
│   └── config.yaml               # 配置文件
└── Makefile                      # 包含 proto 生成命令
```

### Proto 定义

```protobuf
// api/user/v1/user.proto
syntax = "proto3";

package api.user.v1;

option go_package = "github.com/your-org/project/api/user/v1;v1";

// 用户服务
service UserService {
    // 获取用户
    rpc GetUser(GetUserReq) returns (GetUserResp);
    // 创建用户
    rpc CreateUser(CreateUserReq) returns (CreateUserResp);
    // 更新用户
    rpc UpdateUser(UpdateUserReq) returns (UpdateUserResp);
    // 删除用户
    rpc DeleteUser(DeleteUserReq) returns (DeleteUserResp);
    // 列出用户
    rpc ListUsers(ListUsersReq) returns (ListUsersResp);
}

message GetUserReq {
    int64 id = 1;
}

message GetUserResp {
    int64 id = 1;
    string name = 2;
    string email = 3;
    string phone = 4;
    int32 status = 5;
}

message CreateUserReq {
    string name = 1;
    string email = 2;
    string phone = 3;
}

message CreateUserResp {
    int64 id = 1;
}

message UpdateUserReq {
    int64 id = 1;
    string name = 2;
    string email = 3;
    string phone = 4;
}

message UpdateUserResp {
    bool success = 1;
}

message DeleteUserReq {
    int64 id = 1;
}

message DeleteUserResp {
    bool success = 1;
}

message ListUsersReq {
    int32 page = 1;
    int32 page_size = 2;
    string keyword = 3;
}

message ListUsersResp {
    repeated GetUserResp users = 1;
    int32 total = 2;
}
```

### 服务端实现

```go
// internal/grpc/user_grpc.go
package grpc

import (
    "context"

    "github.com/ngq/gorp/framework/contract"
    pb "github.com/your-org/project/api/user/v1"
    "github.com/your-org/project/internal/repository"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// UserServiceServer 用户服务 gRPC 服务端
type UserServiceServer struct {
    pb.UnimplementedUserServiceServer

    userRepo repository.UserRepository
    logger   contract.Logger
}

// NewUserServiceServer 创建用户 gRPC 服务端
func NewUserServiceServer(userRepo repository.UserRepository, logger contract.Logger) *UserServiceServer {
    return &UserServiceServer{
        userRepo: userRepo,
        logger:   logger,
    }
}

// GetUser 获取用户
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserReq) (*pb.GetUserResp, error) {
    s.logger.Info(ctx, "GetUser called", map[string]any{"id": req.Id})

    user, err := s.userRepo.FindByID(ctx, req.Id)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "用户不存在: %d", req.Id)
    }

    return &pb.GetUserResp{
        Id:     user.ID,
        Name:   user.Name,
        Email:  user.Email,
        Phone:  user.Phone,
        Status: int32(user.Status),
    }, nil
}

// CreateUser 创建用户
func (s *UserServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserResp, error) {
    user := &model.User{
        Name:  req.Name,
        Email: req.Email,
        Phone: req.Phone,
    }

    err := s.userRepo.Create(ctx, user)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "创建用户失败: %v", err)
    }

    return &pb.CreateUserResp{Id: user.ID}, nil
}

// ... 其他方法实现
```

### 客户端封装

```go
// internal/client/user_client.go
package client

import (
    "context"

    "github.com/ngq/gorp/framework/contract"
    pb "github.com/your-org/project/api/user/v1"
)

// UserClient 用户服务客户端（类型安全封装）
type UserClient struct {
    rpcClient contract.RPCClient
}

// NewUserClient 创建用户服务客户端
func NewUserClient(rpcClient contract.RPCClient) *UserClient {
    return &UserClient{rpcClient: rpcClient}
}

// GetUser 获取用户
func (c *UserClient) GetUser(ctx context.Context, req *pb.GetUserReq) (*pb.GetUserResp, error) {
    resp := &pb.GetUserResp{}
    err := c.rpcClient.Call(ctx, "user-service", "/api.user.v1.UserService/GetUser", req, resp)
    return resp, err
}

// CreateUser 创建用户
func (c *UserClient) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserResp, error) {
    resp := &pb.CreateUserResp{}
    err := c.rpcClient.Call(ctx, "user-service", "/api.user.v1.UserService/CreateUser", req, resp)
    return resp, err
}

// ListUsers 列出用户
func (c *UserClient) ListUsers(ctx context.Context, req *pb.ListUsersReq) (*pb.ListUsersResp, error) {
    resp := &pb.ListUsersResp{}
    err := c.rpcClient.Call(ctx, "user-service", "/api.user.v1.UserService/ListUsers", req, resp)
    return resp, err
}
```

### 服务启动

```go
// cmd/server/main.go
package main

import (
    "context"

    "github.com/ngq/gorp/app/grpc"
    "github.com/ngq/gorp/framework/bootstrap"
    "github.com/ngq/gorp/framework/container"
    "github.com/ngq/gorp/framework/contract"
    "google.golang.org/grpc"

    pb "github.com/your-org/project/api/user/v1"
    "github.com/your-org/project/internal/grpc"
    "github.com/your-org/project/internal/repository"
)

func main() {
    // 创建服务运行时
    runtime := bootstrap.NewHTTPServiceRuntime(&bootstrap.HTTPServiceOptions{
        Name:    "user-service",
        Address: ":8080",
    })

    c := runtime.Container()

    // 获取依赖
    logger := container.MustMakeLogger(c)
    db := container.MustMakeGorm(c)
    userRepo := repository.NewUserRepository(db)

    // 创建 gRPC Server（使用框架提供的能力）
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(grpc.UnaryServerInterceptor()),
    )

    // 注册 Proto 服务
    userService := grpc.NewUserServiceServer(userRepo, logger)
    pb.RegisterUserServiceServer(grpcServer, userService)

    // 启动 gRPC 服务（独立端口）
    go func() {
        lis, _ := net.Listen("tcp", ":9090")
        grpcServer.Serve(lis)
    }()

    // 启动 HTTP 服务
    bootstrap.RunHTTP(runtime)
}
```

---

## 配置详解

### RPC 配置

```yaml
# config.yaml
rpc:
  # RPC 模式: noop, http, grpc
  mode: grpc

  # 超时配置（毫秒）
  timeout_ms: 30000

  # gRPC 配置
  grpc:
    # 服务端监听地址
    address: ":9090"

    # 目标地址（客户端，无服务发现时使用）
    target: "localhost:9090"

    # 是否使用 insecure（开发环境）
    insecure: true

  # HTTP 配置（mode: http 时使用）
  http:
    base_url: "http://localhost:8080"

  # 服务发现配置
  registry:
    # 注册中心类型: consul, etcd, nacos, noop
    type: consul
    address: "127.0.0.1:8500"
```

### 配置项说明

| 配置项 | 类型 | 说明 | 默认值 |
|-------|------|------|-------|
| `rpc.mode` | string | RPC 模式：noop/http/grpc | grpc |
| `rpc.timeout_ms` | int | 调用超时（毫秒） | 30000 |
| `rpc.grpc.address` | string | gRPC 服务端监听地址 | :9090 |
| `rpc.grpc.target` | string | gRPC 客户端目标地址 | - |
| `rpc.grpc.insecure` | bool | 是否跳过 TLS | true |
| `rpc.http.base_url` | string | HTTP 客户端基础 URL | - |

---

## 拦截器与中间件

### 框架内置拦截器

框架在 `app/grpc/interceptor.go` 中提供了完整的拦截器：

```go
// UnaryServerInterceptor 服务端一元拦截器
// 功能：
// - 从 metadata 提取 TraceID、RequestID
// - 存入 context 供后续使用
// - 收集 Prometheus 指标（请求总数、耗时、当前处理数）
func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        start := time.Now()
        method := info.FullMethod
        grpcRequestsInFlight.WithLabelValues(method).Inc()
        defer grpcRequestsInFlight.WithLabelValues(method).Dec()

        // 从 metadata 提取 TraceID
        md, ok := metadata.FromIncomingContext(ctx)
        if ok {
            if values := md.Get("x-trace-id"); len(values) > 0 {
                ctx = context.WithValue(ctx, "trace_id", values[0])
            }
        }

        resp, err := handler(ctx, req)

        // 记录指标
        st := status.Code(err).String()
        duration := time.Since(start).Seconds()
        grpcRequestsTotal.WithLabelValues(method, st).Inc()
        grpcRequestDuration.WithLabelValues(method, st).Observe(duration)

        return resp, err
    }
}

// UnaryClientInterceptor 客户端一元拦截器
// 功能：
// - 向 metadata 注入 TraceID
// - 实现跨服务链路追踪
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
    return func(ctx context.Context, method string, req, reply interface{},
        cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

        // 从 context 获取 TraceID
        tid := GetTraceID(ctx)

        // 注入到 outgoing metadata
        if tid != "" {
            md := metadata.New(map[string]string{"x-trace-id": tid})
            ctx = metadata.NewOutgoingContext(ctx, md)
        }

        return invoker(ctx, method, req, reply, cc, opts...)
    }
}
```

### Prometheus 指标

框架自动收集的 gRPC 指标：

```go
// gRPC 请求总数
gorp_grpc_requests_total{method="/api.user.v1.UserService/GetUser",status="OK"}

// gRPC 请求耗时
gorp_grpc_request_duration_seconds{method="/api.user.v1.UserService/GetUser",status="OK"}

// 当前处理中的请求数
gorp_grpc_requests_in_flight{method="/api.user.v1.UserService/GetUser"}
```

### 自定义拦截器

```go
// 自定义日志拦截器
func LoggingInterceptor(logger contract.Logger) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // 记录请求
        logger.Info(ctx, "gRPC request", map[string]any{
            "method": info.FullMethod,
            "req":    req,
        })

        resp, err := handler(ctx, req)

        // 记录响应
        if err != nil {
            logger.Error(ctx, "gRPC error", map[string]any{
                "method": info.FullMethod,
                "err":    err.Error(),
            })
        }

        return resp, err
    }
}

// 自定义认证拦截器
func AuthInterceptor(auth contract.ServiceAuthenticator) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // 从 metadata 获取 token
        md, ok := metadata.FromIncomingContext(ctx)
        if !ok {
            return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
        }

        tokens := md.Get("x-service-token")
        if len(tokens) == 0 {
            return nil, status.Errorf(codes.Unauthenticated, "missing token")
        }

        // 验证 token
        if !auth.ValidateToken(ctx, tokens[0]) {
            return nil, status.Errorf(codes.Unauthenticated, "invalid token")
        }

        return handler(ctx, req)
    }
}

// 使用拦截器
func main() {
    srv := grpc.NewServer(
        grpc.ChainUnaryInterceptor(
            grpc.UnaryServerInterceptor(),           // 框架内置
            LoggingInterceptor(logger),              // 自定义日志
            AuthInterceptor(auth),                   // 自定义认证
        ),
    )
}
```

---

## 服务发现集成

### 配置服务发现

```yaml
# config.yaml
rpc:
  mode: grpc
  registry:
    type: consul
    address: "127.0.0.1:8500"

# 或使用 etcd
# registry:
#   type: etcd
#   address: "127.0.0.1:2379"

# 或使用 nacos
# registry:
#   type: nacos
#   address: "127.0.0.1:8848"
```

### 服务注册

框架自动在服务启动时注册：

```go
// framework/provider/rpc/grpc/provider.go 中的服务注册逻辑
func (s *Server) Start(ctx context.Context) error {
    // ...

    // 注册服务到注册中心（如果配置了 Registry）
    if registry != nil {
        serviceName := "user-service"
        addr := s.addr
        err := registry.Register(ctx, serviceName, addr, map[string]string{
            "version": "v1",
        })
    }

    // ...
}
```

### 服务发现流程

```go
// Client.resolveTarget 方法
func (c *Client) resolveTarget(ctx context.Context, service string) (string, contract.DoneFunc, error) {
    // 1. 通过 Registry 发现服务实例
    if c.registry != nil {
        instances, err := c.registry.Discover(ctx, service)
        if err == nil && len(instances) > 0 {
            // 2. 通过 Selector 选择实例（负载均衡）
            if c.selector != nil {
                selected, done, err := c.selector.Select(ctx, instances)
                if err == nil {
                    return selected.Address, done, nil
                }
            }

            // 3. 默认选择第一个健康实例
            for _, inst := range instances {
                if inst.Healthy {
                    return inst.Address, nil, nil
                }
            }
        }
    }

    // 4. 回退到配置的 Target
    if c.cfg.Target != "" {
        return c.cfg.Target, nil, nil
    }

    // 5. 回退到服务名（依赖 DNS）
    return service, nil, nil
}
```

### 负载均衡策略

框架通过 `Selector` 接口支持多种负载均衡策略：

```go
// contract/selector.go
type Selector interface {
    // Select 选择一个服务实例
    Select(ctx context.Context, instances []ServiceInstance) (ServiceInstance, DoneFunc, error)
}
```

常见策略实现：
- **RoundRobin** - 轮询
- **Random** - 随机
- **Weighted** - 权重
- **LeastConnection** - 最少连接

---

## 链路追踪集成

### TraceID 传播

框架自动在 gRPC 调用中传播 TraceID：

**服务端提取**：
```go
// framework/provider/rpc/grpc/provider.go
// 从 incoming metadata 提取 TraceID
md, ok := metadata.FromIncomingContext(ctx)
if ok {
    if values := md.Get("x-trace-id"); len(values) > 0 {
        ctx = context.WithValue(ctx, "trace_id", values[0])
    }
}
```

**客户端注入**：
```go
// framework/provider/rpc/grpc/provider.go
// 准备 outgoing metadata
md := metadata.New(nil)

// 注入 TraceID（如果存在）
if traceID := ctx.Value("trace_id"); traceID != nil {
    md.Set("x-trace-id", fmt.Sprintf("%v", traceID))
}

// 注入 Tracer 上下文（如果配置了 Tracer）
if c.tracer != nil {
    carrier := newGRPCMetadataCarrier(md)
    c.tracer.Inject(ctx, carrier)
}

ctx = metadata.NewOutgoingContext(ctx, md)
```

### 在服务中使用 TraceID

```go
// 获取 TraceID（框架提供的工具方法）
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserReq) (*pb.GetUserResp, error) {
    // 从 context 获取 TraceID
    traceID := appgrpc.GetTraceID(ctx)

    s.logger.Info(ctx, "GetUser called", map[string]any{
        "trace_id": traceID,
        "id":       req.Id,
    })

    // 业务逻辑...

    return resp, nil
}
```

---

## 最佳实践

### 1. Makefile 配置

```makefile
.PHONY: proto proto-gen proto-clean

# 生成 Proto 代码
proto:
	protoc --go_out=. --go-grpc_out=. \
		-I=api \
		--go_opt=paths=source_relative \
		--go-grpc_opt=paths=source_relative \
		api/user/v1/*.proto \
		api/order/v1/*.proto \
		api/inventory/v1/*.proto

# 生成并格式化
proto-gen: proto
	go fmt ./api/...

# 清理生成文件
proto-clean:
	rm -rf api/**/*.pb.go api/**/*_grpc.pb.go

# 完整构建流程
build: proto-gen
	go build ./cmd/server

# 开发流程
dev: proto-gen
	go run ./cmd/server

# 测试
test: proto-gen
	go test ./... -v
```

### 2. 错误处理规范

```go
// internal/grpc/errors.go
package grpc

import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

// 业务错误码映射
var (
    ErrUserNotFound     = status.Errorf(codes.NotFound, "用户不存在")
    ErrUserAlreadyExist = status.Errorf(codes.AlreadyExists, "用户已存在")
    ErrInvalidInput     = status.Errorf(codes.InvalidArgument, "输入参数无效")
    ErrInternal         = status.Errorf(codes.Internal, "内部错误")
)

// 错误处理函数
func HandleError(err error) error {
    if err == nil {
        return nil
    }

    // 根据错误类型映射 gRPC 状态码
    switch {
    case IsNotFoundError(err):
        return status.Errorf(codes.NotFound, err.Error())
    case IsValidationError(err):
        return status.Errorf(codes.InvalidArgument, err.Error())
    case IsDuplicateError(err):
        return status.Errorf(codes.AlreadyExists, err.Error())
    default:
        return status.Errorf(codes.Internal, err.Error())
    }
}

// 在服务中使用
func (s *UserServiceServer) GetUser(ctx context.Context, req *pb.GetUserReq) (*pb.GetUserResp, error) {
    user, err := s.userRepo.FindByID(ctx, req.Id)
    if err != nil {
        return nil, HandleError(err)
    }
    return &pb.GetUserResp{Id: user.ID, Name: user.Name}, nil
}
```

### 3. 请求验证

```go
// 使用 protobuf validate
import "validate/validate.proto";

message CreateUserReq {
    string name = 1 [(validate.rules).string = {min_len: 1, max_len: 100}];
    string email = 2 [(validate.rules).string.email = true];
    string phone = 3 [(validate.rules).string.pattern = "^[0-9]{11}$"];
}

// 或在服务中手动验证
func (s *UserServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserReq) (*pb.CreateUserResp, error) {
    // 验证请求
    if req.Name == "" {
        return nil, status.Errorf(codes.InvalidArgument, "name 不能为空")
    }
    if req.Email == "" {
        return nil, status.Errorf(codes.InvalidArgument, "email 不能为空")
    }

    // 业务逻辑...
}
```

### 4. 单元测试

```go
// internal/grpc/user_grpc_test.go
package grpc

import (
    "context"
    "testing"

    "github.com/ngq/gorp/framework/provider/log"
    pb "github.com/your-org/project/api/user/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

func setupTestServer(t *testing.T) (*grpc.ClientConn, func()) {
    // 创建内存 gRPC 服务器
    lis := bufconn.Listen(bufSize)

    logger := log.NewProvider().Boot(nil)

    srv := grpc.NewServer()
    pb.RegisterUserServiceServer(srv, NewUserServiceServer(nil, logger))

    go srv.Serve(lis)

    // 创建客户端连接
    conn, err := grpc.DialContext(context.Background(), "bufnet",
        grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
            return lis.Dial()
        }),
        grpc.WithInsecure(),
    )
    if err != nil {
        t.Fatalf("Failed to dial: %v", err)
    }

    return conn, func() {
        conn.Close()
        srv.Stop()
    }
}

func TestGetUser(t *testing.T) {
    conn, cleanup := setupTestServer(t)
    defer cleanup()

    client := pb.NewUserServiceClient(conn)

    // 测试正常请求
    resp, err := client.GetUser(context.Background(), &pb.GetUserReq{Id: 1})
    if err != nil {
        t.Fatalf("GetUser failed: %v", err)
    }
    if resp.Name == "" {
        t.Error("Expected non-empty name")
    }

    // 测试错误请求
    _, err = client.GetUser(context.Background(), &pb.GetUserReq{Id: -1})
    if err == nil {
        t.Error("Expected error for invalid id")
    }
}
```

### 5. 性能优化

```go
// 启用压缩（大消息）
import "google.golang.org/grpc/encoding/gzip"

conn, err := grpc.Dial(addr,
    grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
)

// 连接池配置
// 框架已内置连接池，无需额外配置

// 批量操作
message BatchGetUsersReq {
    repeated int64 ids = 1;
}

message BatchGetUsersResp {
    repeated GetUserResp users = 1;
}

// 流式 RPC（大数据量）
service UserService {
    rpc StreamUsers(StreamUsersReq) returns (stream GetUserResp);
}
```

---

## 常见问题

### Q1: 方法名格式是什么？

**gRPC 方法全名格式**：`/{package}.{ServiceName}/{MethodName}`

示例：
- `/api.user.v1.UserService/GetUser`
- `/inventory.InventoryService/ReserveStock`

**查看方法名**：
```bash
# 使用 grpcurl 查看服务方法
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext localhost:9090 describe api.user.v1.UserService
```

### Q2: 如何在框架中注册 Proto 生成的服务？

```go
// 方式1：获取底层 grpc.Server
rpcServer := container.MustMake[contract.RPCServer](c, contract.RPCServerKey)
grpcServer := rpcServer.GRPCServer() // 需要框架暴露此方法
pb.RegisterUserServiceServer(grpcServer, &UserServiceImpl{})

// 方式2：直接创建 grpc.Server（推荐）
func main() {
    // 使用框架提供的拦截器
    grpcServer := grpc.NewServer(
        grpc.UnaryInterceptor(appgrpc.UnaryServerInterceptor()),
    )
    pb.RegisterUserServiceServer(grpcServer, &UserServiceImpl{})

    // 启动 gRPC
    lis, _ := net.Listen("tcp", ":9090")
    go grpcServer.Serve(lis)

    // 启动 HTTP（使用框架）
    bootstrap.RunHTTP(runtime)
}
```

### Q3: 如何处理连接断开重连？

框架内置连接池会自动处理：

```go
// framework/provider/rpc/grpc/provider.go
func (c *Client) getConn(ctx context.Context, service string) (*grpc.ClientConn, contract.DoneFunc, error) {
    // 从缓存获取
    if cached, ok := c.connPool.Load(addr); ok {
        conn := cached.(*grpc.ClientConn)
        // 检查连接状态
        if conn.GetState().String() != "SHUTDOWN" {
            return conn, done, nil
        }
        // 删除无效连接
        c.connPool.Delete(addr)
    }
    // 创建新连接
    // ...
}
```

### Q4: 性能差异大吗？

| 方案 | 序列化 | 连接管理 | 开销 |
|-----|-------|---------|------|
| 框架 Provider | Protobuf | 连接池 | 轻微封装开销 |
| 标准 Proto | Protobuf | 手动管理 | 无额外开销 |

**结论**：框架封装开销可忽略不计，主要耗时在网络 I/O。

### Q5: 如何处理超大消息？

```go
// 启用压缩
import "google.golang.org/grpc/encoding/gzip"

conn, err := grpc.Dial(addr,
    grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
)

// 或使用流式 RPC
service UserService {
    rpc StreamUsers(StreamUsersReq) returns (stream GetUserResp);
}
```

### Q6: 如何调试 gRPC 调用？

```bash
# 安装 grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 查看服务列表
grpcurl -plaintext localhost:9090 list

# 调用方法
grpcurl -plaintext -d '{"id": 123}' localhost:9090 api.user.v1.UserService/GetUser

# 查看服务描述
grpcurl -plaintext localhost:9090 describe api.user.v1.UserService
```

### Q7: 单体项目怎么办？

使用 noop Provider，无需 gRPC：

```yaml
# config.yaml
rpc:
  mode: noop
```

框架会返回错误，单体项目无需服务间调用。

---

## 总结

### 方案选择建议

| 方案 | 开发效率 | 类型安全 | 服务发现 | 框架集成 | 推荐场景 |
|-----|---------|---------|---------|---------|---------|
| 纯框架 Call | ⭐⭐ | ❌ | ✅ | ✅ | 不推荐日常使用 |
| 框架 + 封装层 | ⭐⭐⭐⭐ | ✅ | ✅ | ✅ | 推荐（项目中已采用） |
| Proto + 框架连接池 | ⭐⭐⭐⭐⭐ | ✅ | ✅ | ✅ | **最佳方案**（需扩展框架） |
| 纯 Proto gRPC | ⭐⭐⭐⭐⭐ | ✅ | ❌ | ❌ | 外部系统对接 |
| noop Provider | - | - | - | - | 单体项目 |

### 最终建议

1. **新项目**：使用混合方案（Proto 生成 + 框架连接池）
2. **现有项目**：继续使用封装层方式（项目中已采用）
3. **外部对接**：使用标准 Proto gRPC
4. **单体项目**：使用 noop Provider

### 开发难度评估

| 方案 | 初次学习 | 日常开发 | 维护成本 |
|-----|---------|---------|---------|
| 纯框架 Call | 低 | 高（手动拼接方法名） | 高（运行时错误） |
| 框架 + 封装层 | 中 | 中（需封装每个方法） | 低 |
| Proto + 框架连接池 | 中 | 低（类型安全） | 低 |
| 纯 Proto gRPC | 低 | 低 | 中（需自行集成） |

**结论**：纯框架 Call 方式开发难度较高，但通过简单封装即可解决大部分问题。推荐使用项目中已有的封装方式或混合方案。