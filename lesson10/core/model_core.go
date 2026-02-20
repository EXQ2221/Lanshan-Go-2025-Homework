package core

import (
	"time"

	"gorm.io/gorm"
)

type Role uint8

const (
	RoleNormal Role = 0
	RoleVIP    Role = 1
	RoleAdmin  Role = 2
)

type User struct {
	gorm.Model

	Username     string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	PasswordHash string     `gorm:"size:255;not null" json:"-"`
	TokenVersion int        `gorm:"not null;default:0" json:"-"`
	AvatarURL    string     `gorm:"size:255" json:"avatar_url,omitempty"`
	Profile      string     `gorm:"size:255" json:"profile,omitempty"`
	Role         Role       `gorm:"not null;default:0" json:"role"`
	VIPExpiresAt *time.Time `gorm:"column:vip_expires_at" json:"vip_expires_at,omitempty"`

	Posts         []Post         `gorm:"foreignKey:AuthorID"`
	Comments      []Comment      `gorm:"foreignKey:AuthorID"`
	Reactions     []Reaction     `gorm:"foreignKey:UserID"`
	Favorites     []Favorite     `gorm:"foreignKey:UserID"`
	Activities    []Activity     `gorm:"foreignKey:ActorID"`
	Notifications []Notification `gorm:"foreignKey:UserID"`
	Followers     []UserFollow   `gorm:"foreignKey:FolloweeID"`
	Followees     []UserFollow   `gorm:"foreignKey:FollowerID"`
}

type PostType uint8

const (
	PostArticle  PostType = 1
	PostQuestion PostType = 2
)

type Post struct {
	gorm.Model

	Type      PostType `gorm:"not null;index" json:"type"`
	AuthorID  uint     `gorm:"not null;index" json:"author_id"`
	Title     string   `gorm:"size:200;not null" json:"title"`
	Content   string   `gorm:"type:longtext;not null" json:"content"`
	IsDeleted uint8    `gorm:"not null;default:0;index" json:"-"`
	Status    uint8    `gorm:"not null;default:0;index" json:"status"` // 0发布 1草稿

	Author     User       `gorm:"foreignKey:AuthorID"`
	Comments   []Comment  `gorm:"foreignKey:TargetID"`
	Reactions  []Reaction `gorm:"foreignKey:TargetID"`
	Favorites  []Favorite `gorm:"foreignKey:TargetID"`
	Activities []Activity `gorm:"foreignKey:TargetID"`
	LikeCount  uint       `gorm:"default:0" json:"like_count"`
}

type CommentTargetType uint8

const (
	CommentOnPost     CommentTargetType = 1
	CommentOnQuestion CommentTargetType = 2
	CommentOnComment  CommentTargetType = 3
)

type Comment struct {
	gorm.Model

	TargetType CommentTargetType `gorm:"not null;index:idx_target_created,priority:1" json:"target_type"`
	TargetID   uint              `gorm:"not null;index:idx_target_created,priority:2" json:"target_id"`
	AuthorID   uint              `gorm:"not null;index" json:"author_id"`
	Content    string            `gorm:"type:text;not null" json:"content"`
	IsDeleted  uint8             `gorm:"not null;default:0;index" json:"-"`
	Depth      uint8             `gorm:"not null;default:0" json:"depth"`
	LikeCount  uint              `gorm:"default:0" json:"like_count"`
}

type PostImage struct {
	gorm.Model

	PostID     uint   `gorm:"not null;index" json:"post_id"`
	UploaderID uint   `gorm:"not null;index" json:"uploader_id"`
	URL        string `gorm:"size:512;not null" json:"url"`
}

type UserFollow struct {
	gorm.Model

	FollowerID uint `gorm:"not null;uniqueIndex:uk_pair" json:"follower_id"`
	FolloweeID uint `gorm:"not null;uniqueIndex:uk_pair;index" json:"followee_id"`
}

type QuestionFollow struct {
	gorm.Model

	UserID     uint `gorm:"not null;uniqueIndex:uk_qf" json:"user_id"`
	QuestionID uint `gorm:"not null;uniqueIndex:uk_qf;index" json:"question_id"`
}

type ReactionTargetType uint8

const (
	ReactPost    ReactionTargetType = 1
	ReactAnswer  ReactionTargetType = 2
	ReactComment ReactionTargetType = 3
)

type Reaction struct {
	gorm.Model

	UserID     uint               `gorm:"not null;uniqueIndex:uk_react" json:"user_id"`
	TargetType ReactionTargetType `gorm:"not null;uniqueIndex:uk_react;index" json:"target_type"`
	TargetID   uint               `gorm:"not null;uniqueIndex:uk_react;index" json:"target_id"`
}

type FavoriteTargetType uint8

const (
	FavPost   FavoriteTargetType = 1
	FavAnswer FavoriteTargetType = 2
)

type Favorite struct {
	gorm.Model

	UserID     uint               `gorm:"not null;uniqueIndex:uk_fav" json:"user_id"`
	TargetType FavoriteTargetType `gorm:"not null;uniqueIndex:uk_fav;index" json:"target_type"`
	TargetID   uint               `gorm:"not null;uniqueIndex:uk_fav;index" json:"target_id"`
}

type ActivityAction uint8

const (
	ActPost           ActivityAction = 1
	ActAnswer         ActivityAction = 2
	ActComment        ActivityAction = 3
	ActLike           ActivityAction = 4
	ActFavorite       ActivityAction = 5
	ActFollowUser     ActivityAction = 6
	ActFollowQuestion ActivityAction = 7
)

type ActivityTargetType uint8

const (
	TargetPost     ActivityTargetType = 1
	TargetAnswer   ActivityTargetType = 2
	TargetComment  ActivityTargetType = 3
	TargetUser     ActivityTargetType = 4
	TargetQuestion ActivityTargetType = 5
)

type Activity struct {
	gorm.Model

	ActorID    uint               `gorm:"not null;index" json:"actor_id"`
	Action     ActivityAction     `gorm:"not null" json:"action"`
	TargetType ActivityTargetType `gorm:"not null" json:"target_type"`
	TargetID   uint               `gorm:"not null" json:"target_id"`
}

type Notification struct {
	gorm.Model

	UserID     uint   `gorm:"not null;index" json:"user_id"`
	Type       uint8  `gorm:"not null" json:"type"`
	ActorID    *uint  `json:"actor_id,omitempty"`
	TargetType *uint8 `json:"target_type,omitempty"`
	TargetID   *uint  `json:"target_id,omitempty"`
	Content    string `gorm:"size:255;not null" json:"content"`
	IsRead     uint8  `gorm:"not null;default:0;index" json:"is_read"`
}

type Conversation struct {
	gorm.Model
}

type ConversationMember struct {
	gorm.Model

	ConversationID uint `gorm:"not null;uniqueIndex:uk_cm" json:"conversation_id"`
	UserID         uint `gorm:"not null;uniqueIndex:uk_cm;index" json:"user_id"`
}

type Message struct {
	gorm.Model

	ConversationID uint   `gorm:"not null;index" json:"conversation_id"`
	SenderID       uint   `gorm:"not null;index" json:"sender_id"`
	Content        string `gorm:"type:text;not null" json:"content"`
}
