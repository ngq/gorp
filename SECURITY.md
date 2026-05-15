# Security Policy

本文档描述 gorp 框架的安全策略和漏洞报告流程。

---

## 支持的版本

以下版本正在接受安全更新：

| 版本 | 支持状态 |
|------|---------|
| >= 1.0.0 | ✓ 支持 |
| < 1.0.0 | ✗ 不支持 |

---

## 安全特性

### 1. 依赖安全

- **govulncheck**：每次 CI 运行时检查 Go 依赖漏洞
- **Trivy**：扫描文件系统和容器镜像漏洞
- **Dependency Review**：PR 中自动检查新增依赖

### 2. 容器安全

- **最小化镜像**：基于 alpine，体积约 20MB
- **非 root 用户**：容器默认以 appuser 运行
- **只读文件系统**：可选启用只读根文件系统
- **SBOM**：每次发布自动生成软件物料清单

### 3. 代码安全

- **CodeQL**：静态代码分析，检测安全漏洞
- **golangci-lint**：代码质量检查，包含安全规则

### 4. 供应链安全

- **SBOM 生成**：使用 anchore/sbom-action 生成 SPDX 格式 SBOM
- **镜像签名**：可选启用 cosign 签名
- **依赖锁定**：go.sum 锁定依赖版本

---

## 漏洞报告

### 报告流程

如果您发现安全漏洞，请**不要**在 GitHub Issues 中公开报告。

请通过以下方式报告：

1. **GitHub Security Advisories**（推荐）
   - 访问 https://github.com/ngq/gorp/security/advisories
   - 点击 "Report a vulnerability"
   - 填写漏洞详情

2. **邮件**（备选）
   - 发送邮件至项目维护者
   - 包含漏洞描述、复现步骤、影响范围

### 报告内容

请包含以下信息：

- 漏洞类型（如：注入、XSS、权限绕过）
- 受影响的版本
- 复现步骤
- 潜在影响
- 建议修复方案（如有）

### 响应时间

- **确认**：3 个工作日内确认收到报告
- **评估**：7 个工作日内评估漏洞严重性
- **修复**：根据严重性，14-60 天内发布修复版本

---

## 安全最佳实践

### 使用建议

1. **保持更新**：使用最新版本，及时更新依赖
2. **配置安全**：不要在代码中硬编码敏感信息
3. **网络隔离**：生产环境使用 NetworkPolicy 限制网络访问
4. **最小权限**：容器使用非 root 用户，限制 capabilities

### 配置示例

```yaml
# Pod 安全配置
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    readOnlyRootFilesystem: true
    allowPrivilegeEscalation: false
    capabilities:
      drop:
        - ALL
```

---

## 安全扫描结果

CI 流水线自动运行以下扫描：

| 扫描类型 | 工具 | 频率 |
|---------|------|------|
| 依赖漏洞 | govulncheck | 每次 push/PR |
| 文件系统漏洞 | Trivy | 每次 push/PR |
| 静态分析 | CodeQL | 每次 push/PR |
| 依赖审查 | dependency-review | 每次 PR |

扫描结果可在 GitHub Security tab 查看。
