# micro-auth-demo

一个用于练习微服务拆分的认证示例工程骨架，包含：

- `gateway`：HTTP 入口网关
- `user-service`：用户服务
- `auth-service`：认证与令牌服务
- `biz-service`：业务示例服务
- `shared`：公共工具与响应封装
- `idl`：Thrift 接口定义

当前仓库提供的是可继续扩展的脚手架，`kitex_gen/` 目录预留给代码生成结果。

## 快速开始

1. 启动依赖服务：

```bash
docker compose up --build -d
```

默认已使用 DaoCloud 的 Docker Hub 镜像前缀。
如果你要切换到别的镜像代理，可以在启动前设置：

```bash
DOCKER_HUB_MIRROR=your-mirror/docker.io \
GO_IMAGE=your-mirror/docker.io/library/golang:1.25.2-alpine \
RUNTIME_IMAGE=your-mirror/docker.io/library/alpine:3.20 \
docker compose up --build -d
```

2. 生成 Kitex 代码：

```bash
./scripts/gen.sh
```

3. 分别启动服务：

```bash
./scripts/run_user.sh
./scripts/run_auth.sh
./scripts/run_gateway.sh
```

## 目录说明

目录结构按服务边界拆分，方便后续继续补充：

- 传输协议与契约：`idl/`
- 公共能力：`shared/`
- 网关层：`gateway/`
- 领域服务：`user-service/`、`auth-service/`、`biz-service/`
