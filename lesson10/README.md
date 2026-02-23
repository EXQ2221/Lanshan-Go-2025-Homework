# Lesson10 论坛项目

这是一个前后端分离的论坛应用，后端使用 Go（Gin + Gorm），前端使用 Next.js（AI生成）。

## 功能
- 用户注册 / 登录
- 帖子、评论、点赞
- 关注 / 粉丝
- 通知与未读数
- 收藏与草稿
- 个人资料管理
- 文件上传（头像 / 图片）
- JWT 刷新流程（access + refresh）

## 亮点
- 完整的 JWT 刷新机制，token 过期会自动刷新并重放请求。
- 通知是完整闭环，有未读数，进通知页自动标记已读。
- 关注/粉丝关系和个人主页数据是联动的，不是单独页面。

## 项目结构
- `main.go`: 后端入口
- `internal/`: 路由、handler、中间件
- `core/`: service、model、token 逻辑
- `dao/`: 数据库连接
- `migrations/`: SQL 迁移
- `static/`: 上传资源（如启用）
- `my-forum-app/`: Next.js 前端

## 环境要求
- Go 1.20+（或你本地版本）
- Node.js 18+（或你本地版本）
- MySQL（或 `.env` 配置的数据库）

## 环境变量
在后端根目录创建 `.env`（示例字段）：
```
JWT_SECRET=your_secret
DB_USER=...
DB_PASS=...
DB_HOST=...
DB_PORT=...
DB_NAME=...
```

前端默认请求后端地址：
- `http://localhost:8080`

## 启动后端
在仓库根目录：
```
go run .
```

## 启动前端
```
cd my-forum-app
pnpm install
pnpm dev
```

访问：
- `http://localhost:3000`

## API 说明
- access token 有效期较短，过期后用 refresh token 自动刷新。
- `/refresh` 必须是 public 路由（不能走鉴权中间件）。
- 进入通知页面会标记为已读（如已启用）。
- 详细接口文档见API.md

## License
课程作业 / 学习用途。
