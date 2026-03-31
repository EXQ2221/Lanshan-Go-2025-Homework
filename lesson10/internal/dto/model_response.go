package dto

import (
	"lesson10/internal/model"
	"time"
)

type UserPublicInfo struct {
	ID             uint          `json:"id"`
	Username       string        `json:"username"`
	AvatarURL      string        `json:"avatar_url,omitempty"`
	Profile        string        `json:"profile,omitempty"`
	Role           model.Role    `json:"role"`
	IsVIP          bool          `json:"is_vip"`
	VIPExpiresAt   *time.Time    `json:"vip_expires_at,omitempty"`
	Posts          []PostSummary `json:"posts"`
	PostTotal      int64         `json:"post_total"`
	FollowingCount int64         `json:"following_count"`
	FollowersCount int64         `json:"followers_count"`
	IsFollowed     bool          `json:"is_followed"`
	Page           int           `json:"page"`
	Size           int           `json:"size"`
}

type PostSummary struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Status    uint8     `json:"status"`
}

type GetCommentsResp struct {
	Comments []CommentItem `json:"comments"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	Size     int           `json:"size"`
}

type CommentItem struct {
	ID         uint      `json:"id"`
	AuthorID   uint      `json:"author_id"`
	AuthorName string    `json:"author_name,omitempty"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	Content    string    `json:"content"`
	Depth      uint8     `json:"depth"`
	CreatedAt  time.Time `json:"created_at"`
	LikeCount  int       `json:"like_count"`
	IsLiked    bool      `json:"is_liked"`
}

type PostDetailResp struct {
	ID              uint
	Type            uint8
	AuthorID        uint
	AuthorName      string
	AuthorAvatarURL string
	Title           string
	Content         string `json:"content" binding:"required"`
	Status          uint8  `json:"status"` // 0=发布 1=草稿
	LikeCount       uint
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type PostListItem struct {
	ID              uint
	Type            uint8
	AuthorID        uint
	AuthorName      string
	AuthorAvatarURL string
	Title           string
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time
}
type ToggleReactionResp struct {
	IsLiked   bool `json:"is_liked"`
	LikeCount uint `json:"like_count"`
}

type FollowUserInfo struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	Profile    string `json:"profile,omitempty"`
	IsFollowed bool   `json:"is_followed"` // 当前登录用户是否关注了这个人
}

type NotificationItem struct {
	ID         uint      `json:"id"`
	Type       uint8     `json:"type"`
	ActorID    uint      `json:"actor_id"`
	ActorName  string    `json:"actor_name,omitempty"`
	TargetType uint8     `json:"target_type,omitempty"`
	TargetID   uint      `json:"target_id,omitempty"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

type FavoriteItem struct {
	ID        uint      `json:"id"`
	Type      uint8     `json:"type"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

type UserBasicInfo struct {
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Profile   string `json:"profile,omitempty"`
}

type TokenPair struct {
	AccessToken      string `thrift:"access_token,1" frugal:"1,default,string" json:"access_token"`
	RefreshToken     string `thrift:"refresh_token,2" frugal:"2,default,string" json:"refresh_token"`
	SessionId        string `thrift:"session_id,3" frugal:"3,default,string" json:"session_id"`
	AccessExpiresAt  int64  `thrift:"access_expires_at,4" frugal:"4,default,i64" json:"access_expires_at"`
	RefreshExpiresAt int64  `thrift:"refresh_expires_at,5" frugal:"5,default,i64" json:"refresh_expires_at"`
}

type SessionInfo struct {
	SessionID  string `json:"session_id"`
	DeviceID   string `json:"device_id"`
	DeviceName string `json:"device_name"`
	UserAgent  string `json:"user_agent"`
	LoginIP    string `json:"login_ip"`
	LastIP     string `json:"last_ip"`
	Status     string `json:"status"`
	Current    bool   `json:"current"`
	CreatedAt  int64  `json:"created_at"`
	LastSeenAt int64  `json:"last_seen_at"`
}
