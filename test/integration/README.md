# gorp 深度集成测试

本目录包含 gorp 框架的端到端集成测试，需要真实 backend 服务支持。

## 测试内容

| 测试文件 | 测试内容 | 依赖服务 |
|----------|----------|----------|
| `grpc_test.go` | gRPC 全链路治理 + metadata/tracing 传播 | mock-grpc-backend |
| `etcd_test.go` | etcd 注册发现 + watch 实时更新 | etcd |
| `redis_mq_test.go` | Redis Pub/Sub + 队列消费 + 消费者组 | redis |
| `tracing_test.go` | OTel tracing span 创建/嵌套/错误记录 | jaeger |

## 运行方式

### 方式一：docker-compose（推荐）

```bash
# 启动所有 backend 服务
docker-compose -f docker-compose.test.yml up -d

# 等待服务健康检查完成
docker-compose -f docker-compose.test.yml ps

# 运行集成测试
docker-compose -f docker-compose.test.yml run test-runner

# 清理
docker-compose -f docker-compose.test.yml down
```

### 方式二：本地 backend

如果你本地已有 etcd/Redis/Jaeger：

```bash
# 设置环境变量
export GORP_TEST_ETCD_ADDR=localhost:2379
export GORP_TEST_REDIS_ADDR=localhost:6379
export GORP_TEST_JAEGER_ADDR=localhost:4317

# 启动 mock gRPC backend（可选）
go run ./mock_backend/main.go &
export GORP_TEST_GRPC_BACKEND_ADDR=localhost:50051

# 运行集成测试（跳过 short 测试）
go test -tags=integration -v ./test/integration/...

# 清理 mock backend
pkill -f mock_backend
```

### 方式三：跳过集成测试

```bash
# 仅运行单元测试（跳过需要 backend 的集成测试）
go test -short ./test/integration/...
```

## 目录结构

```
test/integration/
├── docker-compose.test.yml    # 测试环境编排
├── Dockerfile.test            # 测试 runner 构建镜像
├── grpc_test.go               # gRPC 全链路测试
├── etcd_test.go               # etcd 注册发现测试
├── redis_mq_test.go           # Redis MQ 测试
├── tracing_test.go            # tracing 测试
├── README.md                  # 本文档
│
├── mock_backend/              # mock gRPC backend
│   ├── Dockerfile.mock
│   ├── main.go                # mock server 入口
│   ├── grpc_server.go         # server 实现
│
└── pb/                        # 测试 protobuf 定义
    ├── test.proto             # proto 定义
    └── test.pb.go             # 手动实现的 pb（非 protoc 生成）
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `GORP_TEST_ETCD_ADDR` | localhost:2379 | etcd 地址 |
| `GORP_TEST_REDIS_ADDR` | localhost:6379 | Redis 地址 |
| `GORP_TEST_JAEGER_ADDR` | localhost:4317 | Jaeger OTLP gRPC 地址 |
| `GORP_TEST_GRPC_BACKEND_ADDR` | localhost:50051 | Mock gRPC backend 地址 |

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| etcd | 2379 | etcd client port |
| redis | 6379 | Redis port |
| jaeger-ui | 16686 | Jaeger UI |
| jaeger-otlp-grpc | 4317 | OTLP gRPC collector |
| jaeger-otlp-http | 4318 | OTLP HTTP collector |
| mock-grpc-backend | 50051 | Mock gRPC server |

## 验证结果

### Jaeger UI

测试完成后，访问 http://localhost:16686 查看 trace：

1. 选择 service: `test-service-*`
2. 查看 span hierarchy
3. 验证 attributes/status 正确

### etcd

```bash
# 查看注册的服务
etcdctl get "" --prefix --keys-only
```

### Redis

```bash
# 查看队列
redis-cli KEYS "test-*"
```

## 添加新测试

1. 创建新的测试文件 `xxx_test.go`
2. 在 `testing.Short()` 条件中跳过 docker 测试
3. 使用 `getEnvOrDefault()` 获取环境变量
4. 更新 `docker-compose.test.yml` 添加新依赖服务
5. 更新本 README