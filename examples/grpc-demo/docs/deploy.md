# 部署说明

## Docker 开发部署

- `deploy/docker/`
- `deploy/compose/docker-compose.yaml`

本地联调：

```bash
make deploy-local
```

如果希望先统一构建镜像再联调：

```bash
make deploy-local-build
```

说明：
- compose 容器名使用 `grpc-demo-*`
- 本地镜像名使用 `grpc-demo-user-service`、`grpc-demo-order-service`、`grpc-demo-product-service`

## Harbor 推送镜像

```bash
make harbor-push HARBOR_REGISTRY=harbor.example.com HARBOR_NAMESPACE=grpc-demo IMAGE_TAG=v1.0.0
```

会先构建三组项目级服务镜像，再统一 tag 与 push。
默认镜像名：

- `grpc-demo-user-service`
- `grpc-demo-order-service`
- `grpc-demo-product-service`

## Kubernetes 测试 / 预发 / 生产部署

默认附带：

- `deploy/kubernetes/base/`
- `deploy/kubernetes/overlays/dev/`
- `deploy/kubernetes/overlays/staging/`
- `deploy/kubernetes/overlays/prod/`

说明：
- 这里提供的是项目部署资产与环境目录骨架，便于本地联调、预发和生产环境演练。
- 这不代表 `contrib/configsource/kubernetes` 或 `contrib/registry/kubernetes` 已经完成真实后端能力接入。

使用前请先替换：

- `deploy/kubernetes/base/secret.yaml` 中的敏感占位值
- `deploy/kubernetes/overlays/staging/kustomization.yaml` 中的镜像仓库与 tag（镜像名已默认对齐 `grpc-demo-*`）
- `deploy/kubernetes/overlays/prod/kustomization.yaml` 中的镜像仓库与 tag（镜像名已默认对齐 `grpc-demo-*`）
- `deploy/kubernetes/overlays/prod/ingress.yaml` 中的域名

最小验证：

```bash
kubectl kustomize deploy/kubernetes/overlays/dev
kubectl kustomize deploy/kubernetes/overlays/staging
kubectl kustomize deploy/kubernetes/overlays/prod
```
