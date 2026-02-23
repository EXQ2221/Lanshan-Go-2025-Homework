# 接口文档

> 基础地址：`http://localhost:8080`

## 通用说明
- 认证：登录后请在请求头中携带 `Authorization: Bearer <access_token>`。
- 速率限制：服务端启用限流（可能返回 429）。
- 错误响应（通用）：
  - 400 `{"error":"bad_request"}` 或 `{"error":"request format error"}`
  - 401 `{"error":"unauthorized"}` / `{"message":"please login"}`
  - 403 `{"error":"forbidden"}`
  - 409 `{"error":"conflict"}`
  - 500 `{"error":"server_error"}`

## 认证与用户

### 注册
- 方法：`POST /register`
- 权限：无需登录
- 请求体：
```json
{
  "username": "string",
  "password": "string"
}
```
- 返回：
```json
{
  "message": "register success",
  "user_id": 1,
  "username": "xxx"
}
```

### 登录
- 方法：`POST /login`
- 权限：无需登录
- 请求体：
```json
{
  "username": "string",
  "password": "string"
}
```
- 返回：
```json
{
  "message": "login success",
  "user_id": 1,
  "username": "xxx",
  "token": "<access_token>",
  "refresh_token": "<refresh_token>"
}
```

### 刷新令牌
- 方法：`POST /refresh`
- 权限：可选鉴权（不要求 access token）
- 请求体：
```json
{
  "refresh_token": "<refresh_token>"
}
```
- 返回：
```json
{
  "access_token": "<new_access_token>",
  "refresh_token": "<new_refresh_token>"
}
```

### 修改密码
- 方法：`PUT /change_pass`
- 权限：需要登录
- 请求体：
```json
{
  "old_pass": "string",
  "new_pass": "string"
}
```
- 返回：
```json
{
  "ok": true,
  "need relog in": true
}
```

### 更新个人简介
- 方法：`PUT /profile`
- 权限：需要登录
- 请求体：
```json
{
  "profile": "string"
}
```
- 返回：
```json
{ "ok": true }
```

### 上传头像
- 方法：`POST /avatar`
- 权限：需要登录
- 请求：`multipart/form-data`
  - `avatar`: 图片文件（jpg/png/webp，<= 5MB）
- 返回：
```json
{ "avatar_url": "/static/uploads/avatars/xxx.png" }
```

## 帖子

### 发布帖子
- 方法：`POST /posts`
- 权限：需要登录
- 请求体：
```json
{
  "type": 1,
  "title": "string",
  "content": "string",
  "status": 0
}
```
- 返回：
```json
{ "ok": true, "post": { "ID": 1, ... } }
```

### 帖子列表
- 方法：`GET /posts`
- 权限：无需登录
- Query：`page` `size` `type` `keyword`
- 返回：
```json
{
  "list": [ ... ],
  "total": 123,
  "page": 1,
  "page_size": 20
}
```

### 帖子详情
- 方法：`GET /posts/:id`
- 权限：可选鉴权
- 返回：
```json
{
  "ID": 1,
  "Type": 1,
  "AuthorID": 2,
  "AuthorName": "xxx",
  "Title": "...",
  "content": "...",
  "Status": 0,
  "LikeCount": 0,
  "CreatedAt": "...",
  "UpdatedAt": "..."
}
```

### 更新帖子
- 方法：`PUT /posts/:id`
- 权限：需要登录（作者）
- 请求体：
```json
{
  "title": "string",
  "content": "string",
  "status": 0
}
```
- 返回：
```json
{ "ok": true, "message": "update success" }
```

### 删除帖子
- 方法：`DELETE /posts/:id`
- 权限：需要登录（作者或管理员）
- 返回：
```json
{ "message": "delete success" }
```

## 评论

### 发表评论
- 方法：`POST /comments`
- 权限：需要登录
- 请求体：
```json
{
  "target_type": 1,
  "target_id": 1,
  "content": "string"
}
```
- 返回：
```json
{ "message": "post success", "comment": { ... } }
```

### 获取一级评论
- 方法：`GET /posts/comments`
- 权限：无需登录
- Query：`target_type` `target_id` `page` `size`
- 返回：
```json
{ "message": "success", "data": { "comments": [...], "total": 0, "page": 1, "size": 20 } }
```

### 获取评论回复
- 方法：`GET /comments/:parent_id/replies`
- 权限：无需登录
- 返回：
```json
{ "message": "success", "data": { "replies": [...], "total": 0 } }
```

### 删除评论
- 方法：`DELETE /comments/:id`
- 权限：需要登录（作者或管理员）
- 返回：
```json
{ "message": "success" }
```

## 关注

### 关注用户
- 方法：`POST /follow/:id`
- 权限：需要登录
- 返回：
```json
{ "message": "success" }
```

### 取消关注
- 方法：`DELETE /follow/:id`
- 权限：需要登录
- 返回：
```json
{ "message": "success" }
```

### 粉丝列表
- 方法：`GET /users/followers/:id`
- 权限：无需登录（可选鉴权用于 `is_followed`）
- Query：`page` `size`
- 返回：
```json
{ "message": "success", "data": { "users": [...], "total": 0, "page": 1, "size": 20 } }
```

### 关注列表
- 方法：`GET /users/following/:id`
- 权限：无需登录（可选鉴权用于 `is_followed`）
- Query：`page` `size`
- 返回：
```json
{ "message": "success", "data": { "users": [...], "total": 0, "page": 1, "size": 20 } }
```

## 用户资料

### 获取用户公开信息
- 方法：`GET /user/:id`
- 权限：可选鉴权
- Query：`page`
- 返回：
```json
{
  "message": "success",
  "data": {
    "id": 1,
    "username": "xxx",
    "avatar_url": "...",
    "profile": "...",
    "role": 0,
    "is_vip": false,
    "vip_expires_at": null,
    "posts": [ ... ],
    "post_total": 0,
    "following_count": 0,
    "followers_count": 0,
    "is_followed": false,
    "page": 1,
    "size": 5
  }
}
```

## 上传

### 上传文章图片
- 方法：`POST /upload/article-image`
- 权限：需要登录
- 请求：`multipart/form-data`
  - `image`: 图片文件（jpg/png/webp，<= 10MB）
- 返回：
```json
{ "message": "success", "image_url": "/static/uploads/images/xxx.png" }
```

## 点赞 / 收藏

### 点赞 / 取消点赞
- 方法：`POST /reactions`
- 权限：需要登录
- 请求体：
```json
{ "target_type": 1, "target_id": 1 }
```
- 返回：
```json
{ "message": "success", "status": true }
```

### 收藏 / 取消收藏
- 方法：`POST /favorites`
- 权限：需要登录
- 请求体：
```json
{ "target_type": 1, "target_id": 1 }
```
- 返回：
```json
{ "message": "success", "data": { "is_favorited": true } }
```

### 收藏列表
- 方法：`GET /favorites`
- 权限：需要登录
- Query：`page` `size`
- 返回：
```json
{ "message": "success", "data": { "favorites": [...], "total": 0, "page": 1, "size": 20 } }
```

### 草稿列表
- 方法：`GET /draft`
- 权限：需要登录
- Query：`page` `size`
- 返回：
```json
{ "message": "success", "data": { "drafts": [...], "total": 0, "page": 1, "size": 20 } }
```

## 通知

### 通知列表
- 方法：`GET /notifications`
- 权限：需要登录
- Query：`page` `size` `unread_only` (0/1)
- 返回：
```json
{ "message": "success", "data": { "notifications": [...], "total": 0, "page": 1, "size": 20 } }
```

### 未读数
- 方法：`GET /notifications/count`
- 权限：需要登录
- 返回：
```json
{ "count": 0 }
```

### 全部标为已读
- 方法：`POST /notifications/read-all`
- 权限：需要登录
- 返回：
```json
{ "message": "success" }
```
