# contrib/registry

当前目录用于承接从 `framework/provider/discovery/*` 逐步外移出来的具体注册中心后端实现。

当前阶段说明：

- `framework/provider/discovery/*` 仍是实际运行实现来源；
- 这里先建立目录骨架，作为 Phase 4 第一批外移的目标落点；
- 真正的后端代码迁移应分能力域逐个推进，并同步测试、引用关系与文档口径。
