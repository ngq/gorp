# 部署说明

## Docker 开发部署
- 构建镜像：`docker build -t monolith:latest .`

## Kubernetes 测试 / 预发部署

默认附带最小 Kubernetes 部署骨架：
- `deploy/kubernetes/base/`
- `deploy/kubernetes/overlays/dev/`
- `deploy/kubernetes/overlays/staging/`

说明：
- 这里提供的是项目部署样板与目录骨架，便于本地演练和环境落地。
- 这不代表 `contrib/configsource/kubernetes` 或 `contrib/registry/kubernetes` 已经完成真实后端能力接入。

使用前请先替换：

- `deploy/kubernetes/base/secret.yaml` 中的敏感占位值
- `deploy/kubernetes/overlays/staging/kustomization.yaml` 中的镜像仓库与 tag（镜像名默认对齐 `monolith`）

默认资源命名：

- ConfigMap：`monolith-config`
- Secret：`monolith-secret`
- Namespace（dev / staging）：`monolith-dev` / `monolith-staging`
- 镜像名：`monolith`

最小验证：

```bash
kubectl kustomize deploy/kubernetes/overlays/dev
kubectl kustomize deploy/kubernetes/overlays/staging
```
