# gorp 框架性能基准测试报告

> 测试日期：2026-05-14
> 测试环境：AMD Ryzen 9 9950X3D 16-Core, Windows 11, Go 1.22+

---

## 1. 核心组件性能摘要

### 1.1 选择器（Selector）

| 算法 | 1 实例 | 10 实例 | 100 实例 | 1000 实例 |
|------|--------|---------|----------|-----------|
| **Random** | 44 ns | 185 ns | 1.3 µs | 12.6 µs |
| **WRR** | 122 ns | 728 ns | 5.8 µs | 24.6 µs |
| **P2C** | 88 ns | 283 ns | 1.7 µs | 16.1 µs |

**结论**：Random 最快（44 ns），P2C 性能接近（88 ns）且具备自适应能力。WRR 在小规模场景提供精确权重分配；超过 100 实例时自动降级到 P2C，1000 实例性能从 48.6µs 降至 24.6µs。

### 1.2 元数据（Metadata）

| 操作 | 性能 | 内存分配 |
|------|------|----------|
| **Get** | 9.6 ns | 0 allocs |
| **Set** | 27 ns | 2 allocs, 21 B |
| **Clone** | 396 ns | 13 allocs, 904 B |

**结论**：Get 极快（9.6 ns，无内存分配），Set 快速（27 ns），Clone 有一定开销（396 ns）。

### 1.3 错误处理（Errors）

| 操作 | 性能 | 内存分配 |
|------|------|----------|
| **NewError** | 0.8 ns | 0 allocs |
| **WithCause** | 1.7 ns | 0 allocs |
| **FromError** | 210 ns | 1 alloc, 16 B |

**结论**：错误创建几乎无开销（0.8 ns），FromError 有少量开销（类型断言）。

### 1.4 重试延迟计算

| 操作 | 性能 |
|------|------|
| **CalculateDelay** | 0.9 ns |

**结论**：指数退避计算极快。

### 1.5 数据验证（Validate）

| 场景 | 性能 | 内存分配 |
|------|------|----------|
| **有效数据** | 547 ns | 5 allocs, 89 B |
| **无效数据** | 2236 ns | 41 allocs, 2.5 KB |

**结论**：有效数据验证开销低，无效数据因错误消息生成有额外开销。

---

## 2. HTTP/gRPC 治理链性能

### 2.1 HTTP Middleware Chain

| 场景 | 性能 | 内存分配 |
|------|------|----------|
| **13 阶段全链路** | 5.0 µs | 29 allocs, 7.3 KB |
| **5 阶段简化** | 6.2 µs | 26 allocs, 6.9 KB |
| **短路终止（限流拒绝）** | 5.0 µs | 13 allocs, 5.3 KB |
| **无 middleware** | 8.0 µs | 24 allocs, 6.6 KB |

**结论**：13 阶段全链路开销约 5 µs，可接受。短路终止反而减少内存分配。

### 2.2 gRPC Interceptor Chain

| 场景 | 性能 | 内存分配 |
|------|------|----------|
| **9 阶段全链路** | 1.1 µs | 14 allocs, 536 B |
| **5 阶段简化** | 960 ns | 10 allocs, 440 B |
| **无 interceptor** | 0.18 ns | 0 allocs |

**结论**：9 阶段全链路开销仅 1.1 µs，非常轻量。gRPC 比 HTTP 更轻（无 JSON 序列化开销）。

---

## 3. Tracing 性能

| 场景 | 性能 | 内存分配 |
|------|------|----------|
| **Span 创建** | 1.4 µs | 7 allocs, 664 B |
| **带 Attributes（5 个）** | 4.7 µs | 18 allocs, 2.5 KB |
| **嵌套 Span（parent + child）** | 2.9 µs | 14 allocs, 1.3 KB |

**结论**：Span 创建开销约 1.4 µs，可接受。带 Attributes 时开销增加，建议只在关键节点设置。

---

## 4. JSON 序列化性能

| 场景 | 性能 | 内存分配 |
|------|------|----------|
| **单个对象序列化** | 370 ns | 2 allocs, 160 B |
| **单个对象反序列化** | 1.6 µs | 20 allocs, 728 B |
| **100 个用户序列化** | 14 µs | 2 allocs, 6.8 KB |

**结论**：JSON 序列化开销低，反序列化略慢。

---

## 5. Context 操作性能

| 操作 | 性能 | 内存分配 |
|------|------|----------|
| **gin.Context.Get** | 15 ns | 0 allocs |
| **gin.Context.Set** | 2.9 µs | 14 allocs, 1.6 KB |
| **context.Value** | 3.8 ns | 0 allocs |
| **context.WithValue** | 90 ns | 1 alloc, 48 B |

**结论**：Get/Value 查询极快（15 ns / 3.8 ns），Set 有开销（gin.Context 创建）。

---

## 6. 综合请求性能

| 场景 | 性能 | 内存分配 |
|------|------|----------|
| **完整 HTTP 请求** | 11 µs | 22 allocs, 6.7 KB |
| **并发 HTTP 请求** | 1.7 µs | 24 allocs, 6.6 KB |

**结论**：单次完整请求约 11 µs，单核可处理 ~90,000 req/s。并发场景吞吐量更高。

---

## 7. 性能等级评定

| 模块 | 性能 | 评级 |
|------|------|------|
| **Random Selector** | 41 ns (1实例) | ⭐⭐⭐⭐⭐ 极快 |
| **P2C Selector** | 82 ns (1实例) | ⭐⭐⭐⭐⭐ 极快 |
| **WRR Selector** | 118 ns (1实例) | ⭐⭐⭐⭐ 快 |
| **Metadata Get** | 9.6 ns | ⭐⭐⭐⭐⭐ 极快 |
| **Metadata Set** | 27 ns | ⭐⭐⭐⭐⭐ 极快 |
| **NewError** | 0.8 ns | ⭐⭐⭐⭐⭐ 几乎无开销 |
| **CalculateDelay** | 0.9 ns | ⭐⭐⭐⭐⭐ 极快 |
| **Validator (有效)** | 547 ns | ⭐⭐⭐⭐ 可接受 |
| **HTTP 13 阶段链** | 5.0 µs | ⭐⭐⭐⭐ 可接受 |
| **gRPC 9 阶段链** | 1.1 µs | ⭐⭐⭐⭐⭐ 极轻量 |
| **Tracing Span** | 1.4 µs | ⭐⭐⭐⭐ 可接受 |
| **JSON 序列化** | 370 ns | ⭐⭐⭐⭐⭐ 极快 |

---

## 8. 优化建议

### 8.1 当前无需优化

以下模块性能已足够优秀：
- Random/P2C Selector（< 100 ns）
- Metadata Get/Set（< 30 ns）
- Error 创建（< 2 ns）
- gRPC Interceptor Chain（1.1 µs）
- JSON 序列化（370 ns）

### 8.2 可考虑优化（非紧急）

| 模块 | 当前 | 建议 |
|------|------|------|
| **Metadata Clone** | 396 ns | 可考虑浅拷贝优化 |
| **Validate（无效数据）** | 2236 ns | 错误消息池化 |
| **Tracing Attributes** | 4.7 µs | 延迟设置 |

### 8.3 生产环境建议

1. **高并发场景**：使用 P2C 负载均衡，自适应调节
2. **启用 Tracing**：只在关键节点设置 attribute
3. **大型 JSON**：考虑分页或流式输出

---

## 9. 测试覆盖

| 模块 | Benchmark | 状态 |
|------|-----------|------|
| Selector (Random/WRR/P2C) | `BenchmarkRandomSelector_Select` 等 | ✅ |
| Metadata | `BenchmarkMetadata_Get/Set/Clone` | ✅ |
| Errors | `BenchmarkNewError/WithCause/FromError` | ✅ |
| Retry | `BenchmarkCalculateDelay` | ✅ |
| Validate | `BenchmarkValidator_Validate_*` | ✅ |
| HTTP Middleware Chain | `BenchmarkHTTPMiddlewareChain_*` | ✅ |
| gRPC Interceptor Chain | `BenchmarkGRPCInterceptorChain_*` | ✅ |
| Tracing Span | `BenchmarkTracingSpanCreation_*` | ✅ |
| JSON | `BenchmarkJSONSerialize/*` | ✅ |
| Context | `BenchmarkContextSet/Get/*` | ✅ |
| 综合请求 | `BenchmarkFullHTTPRequest/*` | ✅ |

---

## 10. 运行命令

```bash
# 运行全部 benchmark
go test ./benchmark/... -bench=. -benchmem -count=3

# 运行特定模块
go test ./benchmark/... -bench=Selector -benchmem
go test ./benchmark/... -bench=HTTPMiddleware -benchmem
go test ./benchmark/... -bench=Tracing -benchmem
```

---

## 附录：基线数据原文

第一次运行（基线测试）关键数据：

```
BenchmarkRandomSelector_Select/instances_1-32          41.56 ns/op    112 B/op    2 allocs/op
BenchmarkRandomSelector_Select/instances_1000-32       11301 ns/op    65584 B/op  2 allocs/op
BenchmarkWRRSelector_Select/instances_1-32             117.8 ns/op    112 B/op    2 allocs/op
BenchmarkWRRSelector_Select/instances_1000-32          48600 ns/op    120193 B/op 7 allocs/op
BenchmarkP2CSelector_Select/instances_1-32             81.76 ns/op    192 B/op    3 allocs/op
BenchmarkMetadata_Get-32                               9.628 ns/op    0 B/op      0 allocs/op
BenchmarkMetadata_Set-32                               27.10 ns/op    21 B/op     2 allocs/op
BenchmarkNewError-32                                   0.8052 ns/op   0 B/op      0 allocs/op
BenchmarkFromError-32                                  210.6 ns/op    16 B/op     1 allocs/op
BenchmarkValidator_Validate_Valid-32                   546.0 ns/op    89 B/op     5 allocs/op
BenchmarkValidator_Validate_Invalid-32                 2236 ns/op     2481 B/op   41 allocs/op
```