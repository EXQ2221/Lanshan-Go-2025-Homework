# Lesson10 论坛项目

这是一个前后端分离的论坛应用：
- 后端：Go（Gin + Gorm）
- 前端：Next.js + TypeScript
- 数据库：MySQL

## 功能
- 用户注册 / 登录
- JWT 鉴权（Access Token + Refresh Token）
- 帖子发布、编辑、删除、详情、列表检索
- 评论与多级回复
- 点赞与收藏
- 关注 / 取关、粉丝 / 关注列表
- 通知系统（未读数、全部已读）
- 个人主页与资料编辑
- 头像与文章图片上传

## 技术实现说明
- 后端采用分层结构：`router -> handler -> service -> repository`。
- 鉴权采用 JWT：
  - Access Token 过期后，前端通过 `/refresh` 携带 `refresh_token` 自动刷新。
  - 刷新时校验 `type/token_version/exp`，并执行 `token_version + 1` 防重放。
- 前端使用 Axios 拦截器：
  - 请求自动携带 `Authorization`。
  - 401 自动刷新并重放原请求。
  - 429 单独提示“操作过于频繁”。
- 帖子检索支持关键词查询，后端基于 MySQL 全文索引（已接入 ngram 方案）。

## 项目亮点
- 完整的 Token 刷新闭环：登录返回双 Token，过期自动刷新并重放请求。
- 通知闭环：触发通知、未读数统计、通知页批量标记已读。
- 用户关系联动：主页、关注、粉丝、是否已关注状态统一。

## 项目结构
- `cmd/server/main.go`：后端启动入口
- `internal/router/`：路由注册
- `internal/handler/`：HTTP 处理层
- `internal/service/`：业务逻辑层
- `internal/repository/`：数据访问层
- `internal/model/`：数据模型
- `internal/dto/`：请求/响应结构
- `internal/middleware/`：鉴权与限流中间件
- `internal/pkg/token/`：JWT 生成与校验
- `migrations/`：SQL 迁移脚本
- `static/`：静态资源与上传文件
- `my-forum-app/`：前端项目
- `API.md`：接口文档

## 环境要求
- Go 1.20+
- Node.js 18+
- pnpm 8+
- MySQL 8+

## 环境变量
后端默认读取 `.env.local` 和 `.env`，至少需要：

```env
APP_PORT=8080
JWT_SECRET=your_secret
JWT_EXPIRE_HOURS=24

DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=lesson10
DB_USER=lesson10_user
DB_PASS=your_password
```

## 启动后端
在项目根目录执行：

```bash
go run ./cmd/server/main.go
```

默认监听：`http://localhost:8080`

## 启动前端

```bash
cd my-forum-app
pnpm install
pnpm dev
```

默认访问：`http://localhost:3000`

## API 说明
- 详细接口见 `API.md`。
- 典型认证流程：
  1. 登录获取 `token + refresh_token`
  2. Access Token 过期后调用 `POST /refresh`
  3. 刷新成功后重放原请求

## License
课程作业 / 学习用途。