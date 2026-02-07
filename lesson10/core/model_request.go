package core

import "time"

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChangePassRequest struct {
	OldPass string `json:"old_pass"`
	NewPass string `json:"new_pass"`
}

type UpdateProfileRequest struct {
	Profile *string `json:"profile" binding:"omitempty,max=255"`
}

type CreatePostRequest struct {
	Type    uint8  `json:"type" binding:"required,oneof=1 2"` // 1/2
	Title   string `json:"title" binding:"required,max=200"`
	Content string `json:"content" binding:"required"` // markdown 字符串
}

type ListPostsQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`        // 当前页码
	PageSize int    `form:"size" binding:"omitempty,min=1,max=50"` // 每页数量，默认 20
	Type     uint8  `form:"type" binding:"omitempty"`              // 帖子类型
	Keyword  string `form:"keyword" binding:"omitempty"`
}
type PostListItem struct {
	ID         uint
	Type       uint8
	AuthorID   uint
	AuthorName string
	Title      string
	CreateAt   time.Time
	UpdatedAt  time.Time
}

type PostDetailResp struct {
	ID         uint
	Type       uint8
	AuthorID   uint
	AuthorName string
	Title      string
	Content    string `json:"content" binding:"required"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PostCommentRequest struct {
	TargetType CommentTargetType `json:"target_type" binding:"required,oneof=1 2 3"`
	TargetID   uint              `json:"target_id" biding:"required"`
	Content    string            `json:"content"`
}

// GetCommentsReq 查询参数
type GetCommentsReq struct {
	TargetType uint8 `form:"target_type" binding:"required,oneof=1 2"`
	TargetID   uint  `form:"target_id" binding:"required"`
	Page       int   `form:"page" default:"1"`
	Size       int   `form:"size" default:"20"`
}

// GetCommentsResp 返回结构
type GetCommentsResp struct {
	Comments []CommentItem `json:"comments"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	Size     int           `json:"size"`
}

type CommentItem struct {
	ID         uint      `json:"id"`
	AuthorID   uint      `json:"author_id"`
	AuthorName string    `json:"author_name,omitempty"` // 可选，联表查用户名
	Content    string    `json:"content"`
	Depth      uint8     `json:"depth"`
	CreatedAt  time.Time `json:"created_at"`
	LikeCount  int       `json:"like_count"`
}

type UpdatePostRequest struct {
	Title   string `json:"title" binding:"omitempty"`
	Content string `json:"content" binding:"omitempty"`
}
