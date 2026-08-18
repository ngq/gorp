# gorp 框架 · 交接总入口（零上下文接手者从这里开始）

> **这是唯一需要"从头读完"的文档**。读完它你就能回答"这是什么项目、做了什么、还差什么、该读哪份文档"。
> 其余文档按需深入（见 §5 文档地图）。

---

## 1. 项目一句话

**gorp**（`github.com/ngq/gorp`，Go 1.26）是一个基于 **Gin** 的微服务框架：单机用 gin 增强（开箱即用），上微服务用 proto-first + 治理（熔断/限流/重试/服务认证/发现），全部通过 DI 容器 + provider 装配，目标是**单机到微服务无缝衔接**。

- **产品定位**（三层）：gin 为脸（对外默认编程模型）、proto 为门（微服务入口）、契约为地基（provider 接口/RPC/生命周期）。详见 `CODE_REVIEW_2026-08-17.zh-CN.md` 第 8 节。
- **结构**：根 module `github.com/ngq/gorp`（framework/ 契约+provider+bootstrap）+ `contrib/` 下 24 个嵌套 module（各中间件/注册中心，独立 go.mod，按需引入）+ `examples/`（被 gitignore，不入库）。

---

## 2. 当前状态快照（2026-08-18）

| 项 | 状态 |
| --- | --- |
| 代码质量 | 两轮深度 review 81 项发现，**P0/P1 全部修复**；`go build/vet/test`（framework+24 嵌套 module）全绿，`-race` 无竞态 |
| 集成验证 | 已做 docker 真实验证（ServiceCenter/Kafka/RabbitMQ/Redis/ZooKeeper），修复 5 个真实缺陷，验证工具保留在 `contrib/*/internal/*verify` |
| 架构 | 契约层已 gRPC-free（依赖倒置完成）；capability fail-fast 已生效；mTLS 接线、内置真实 tracer、MQ 模型文档化均完成 |
| 二进制 | HTTP-only 42MB（stripped）→ 目标 32MB，**剩最后一步包拆分未做**（有开发文档） |
| 工作树 | 干净（除 `.idea/golinter.xml`、`CODE_REVIEW_2026-07-20.zh-CN.md` 两个会话前无关改动） |

---

## 3. 已做的工作（一句话摘要 + 依据）

| 主题 | 内容 | 依据 |
| --- | --- | --- |
| 安全 | HTTP 认证 fail-closed、SSH 主机密钥校验、JWT issuer/audience 校验、mTLS 传输层接线、类型化 context key | commit `71fdab0`/`3fd4fd3` |
| 正确性 | 81 项 review 发现修复（ETag 吞响应、config 死锁、事件 panic、Prometheus 标签错位、noop 锁失效、trace 断裂等） | commit `8e95d0a`/`1c4d9d7`/`9f17aa3`/`fa4bed9` |
| 集成 | 5 个中间件/注册中心真实缺陷修复 + 回归验证工具 | commit `8221417` |
| 决策 | capability 未知后端 fail-fast；MQ 广播/队列模型文档化 | commit `75f6ece`/`a3a745d` |
| 架构 | gRPC 契约解耦基础层（契约 grpc-free + retry 本地分类 + 适配层） | commit `4ceb756`/`80d9387` |

---

## 4. 未做/待办（复核时勿误报为 bug）

| 项 | 状态 | 文档 |
| --- | --- | --- |
| 二进制瘦身最后一步（grpc 包拆分，42→32MB） | **未做，有开发文档** | `HANDOFF_2026-08-18.gRPC包拆分开发任务.zh-CN.md` |
| BBR-02 非 Linux CPU 采样、REG-06 ServiceCenter 增强（ak/sk 等） | 未做，需外部环境 | `HANDOFF_2026-08-18.gRPC解耦方案与待办清单.zh-CN.md` §B |
| DUP-01/02 ginContext 重复 | 未做，依赖方向重构 | 同上 |
| CHANGELOG.md 提交 | 内容已补记，被本地 exclude 挡住 | — |
| WS-ROOM-01 反向索引 | **已判定不做**（有记录） | 同上 §B.4 |

---

## 5. 文档地图（按需深入）

| 想了解 | 读 |
| --- | --- |
| **项目架构/产品定位/81 项修复详情** | `CODE_REVIEW_2026-08-17.zh-CN.md`（第 8 节=定位，第 10 节=修复落地状态） |
| **第一轮修复约定** | `CODE_REVIEW_2026-08-16.zh-CN.md` 第 10 节 |
| **全量变更摘要 + Breaking changes** | `CHANGELOG.md` |
| **所有已改项的复核要点**（复核用） | `HANDOFF_2026-08-18.全量变更复核清单.zh-CN.md` |
| **gRPC 解耦方案 + 剩余待办全量** | `HANDOFF_2026-08-18.gRPC解耦方案与待办清单.zh-CN.md` |
| **下一步开发：grpc 包拆分**（要动手做时读） | `HANDOFF_2026-08-18.gRPC包拆分开发任务.zh-CN.md` |
| **如何 build/test** | `README.zh-CN.md` |
| **代码风格/提交约定** | `CONTRIBUTING.md` |

**推荐阅读顺序**：
1. 本文档（全景）
2. `CODE_REVIEW_2026-08-17.zh-CN.md` 第 8+10 节（架构 + 修复约定）
3. 按任务选：复核 → 复核清单；开发 → 包拆分任务；理解待办 → 方案与待办清单

---

## 6. 环境警告（防踩坑）

- `CHANGELOG.md` 未纳入 git（`.git/info/exclude` 本地排除），内容有效。
- `examples/` 整个目录被 gitignore（未入库），改动仅本地验证。
- `.idea/golinter.xml`、`CODE_REVIEW_2026-07-20.zh-CN.md` 是会话前无关改动，勿动勿提交。
- 体积测试必须 `-ldflags="-s -w"`（plain 89MB 会误导）。
- contrib 是 24 个嵌套 module，build/test 需各自 `cd`。
