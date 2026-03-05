package service

import (
	"context"
	"errors"
	"lesson10/internal/config"
	"lesson10/internal/dto"
	"lesson10/internal/model"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/repository"
	"strings"
	"time"

	"gorm.io/gorm"
)

type PostService struct {
	userRepo     repository.UserRepository
	postRepo     repository.PostRepository
	favoriteRepo repository.FavoriteRepository
}

func NewPostService(userRepo repository.UserRepository, postRepo repository.PostRepository, favoriteRepo repository.FavoriteRepository) *PostService {
	return &PostService{
		userRepo:     userRepo,
		postRepo:     postRepo,
		favoriteRepo: favoriteRepo,
	}
}

func (r *PostService) CreatePostService(ctx context.Context, req *dto.CreatePostRequest, authorID uint) (*model.Post, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, errcode.ErrBadRequest
	}

	if strings.TrimSpace(req.Content) == "" {
		return nil, errcode.ErrBadRequest
	}

	p := &model.Post{
		Type:     model.PostType(req.Type),
		AuthorID: authorID,
		Title:    title,
		Content:  req.Content,
		Status:   req.Status,
	}

	if err := r.postRepo.CreatePost(ctx, p); err != nil {
		return nil, errcode.ErrInternal
	}

	return p, nil
}

func ListPostsService(q dto.ListPostsQuery) ([]dto.PostListItem, int64, error) {
	page := q.Page
	if page == 0 {
		page = 1
	}
	ps := q.PageSize
	if ps == 0 {
		ps = 20
	}
	offset := (page - 1) * ps

	// 基础查询：只查未删除、已发布的帖子
	db := config.DB.Model(&model.Post{}).
		Where("is_deleted = 0 AND status = 0")

	if q.Type > 0 {
		db = db.Where("type = ?", q.Type)
	}

	// 关键词搜索（标题或内容）
	if q.Keyword != "" {
		keyword := "%" + q.Keyword + "%"
		db = db.Where("title LIKE ? OR content LIKE ?", keyword, keyword)
	}

	// 统计总数（不变）
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, errcode.ErrInternal
	}

	// 改用 JOIN 查询，同时取出作者用户名
	// 注意：这里使用 Table + Joins + Select，性能更好，也能避免 Preload 的 N+1 问题
	var rows []struct {
		ID         uint      `gorm:"column:id"`
		Type       uint8     `gorm:"column:type"`
		AuthorID   uint      `gorm:"column:author_id"`
		Title      string    `gorm:"column:title"`
		CreatedAt  time.Time `gorm:"column:created_at"`
		UpdatedAt  time.Time `gorm:"column:updated_at"`
		AuthorName string    `gorm:"column:username"` // 从 users 表取
	}

	err := db.Table("posts p").
		Joins("LEFT JOIN users u ON p.author_id = u.id").
		Select("p.id, p.type, p.author_id, p.title, p.created_at, p.updated_at, u.username").
		Order("p.updated_at DESC"). // ← 改用 updated_at 排序（最新活跃的排前面）
		Limit(ps).
		Offset(offset).
		Scan(&rows).Error

	if err != nil {
		return nil, 0, errcode.ErrInternal
	}

	// 转换成前端需要的 PostListItem 结构
	items := make([]dto.PostListItem, len(rows))
	for i, r := range rows {
		items[i] = dto.PostListItem{
			ID:         r.ID,
			Type:       r.Type,
			AuthorID:   r.AuthorID,
			AuthorName: r.AuthorName, // ← 这里就有用户名了
			Title:      r.Title,
			CreateAt:   r.CreatedAt,
			UpdatedAt:  r.UpdatedAt,
		}
	}

	return items, total, nil
}
func (r *PostService) GetPostService(ctx context.Context, currentID, id uint) (*dto.PostDetailResp, error) {
	var p model.Post
	if err := r.postRepo.FindPostByID(ctx, id, &p); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal
	}

	if p.Status == 1 {
		if p.AuthorID != currentID {
			return nil, errcode.ErrUnauthorized
		}
	}

	return &dto.PostDetailResp{
		ID:         p.ID,
		Type:       uint8(p.Type),
		AuthorID:   p.AuthorID,
		AuthorName: p.Author.Username,
		Title:      p.Title,
		Content:    p.Content,
		Status:     p.Status,
		LikeCount:  p.LikeCount,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}, nil
}

func (r *PostService) UpdatePostService(ctx context.Context, PostID uint64, id uint, req *dto.UpdatePostRequest) error {
	var post model.Post
	err := r.postRepo.FindPostByID(ctx, uint(PostID), &post)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrNotFound
	}
	if err != nil {
		return errcode.ErrInternal
	}

	if post.AuthorID != id {
		return errcode.ErrUnauthorized
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = strings.TrimSpace(req.Title)
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	updates["updated_at"] = time.Now()

	return r.postRepo.UpdatePost(ctx, updates, post)
}

func (r *PostService) DeletePostService(ctx context.Context, postID, uid uint, role uint) error {
	var post model.Post
	err := r.postRepo.FindPostByID(ctx, postID, &post)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrNotFound
	}
	if err != nil {
		return errcode.ErrInternal
	}

	if post.AuthorID != uid && role != uint(model.RoleAdmin) {
		return errcode.ErrUnauthorized
	}

	return r.postRepo.DeletePost(ctx, post)

}

func (r *PostService) GetFavoritesService(ctx context.Context, uid uint, page, size int) ([]dto.FavoriteItem, int64, error) {
	offset := (page - 1) * size

	total, err := r.favoriteRepo.CountByUserID(ctx, uid)

	var favorites []model.Favorite
	err = r.favoriteRepo.ListByUserID(ctx, uid, offset, size, &favorites)

	if err != nil {
		return nil, 0, err
	}

	// 批量查收藏的帖子信息（从 posts 表）
	postIDs := make([]uint, 0, len(favorites))
	for _, f := range favorites {
		postIDs = append(postIDs, f.TargetID)
	}

	var posts []model.Post
	if len(postIDs) > 0 {
		r.postRepo.FindPostsByIDs(ctx, postIDs, &posts)
	}

	postMap := make(map[uint]model.Post)
	for _, p := range posts {
		postMap[p.ID] = p
	}

	items := make([]dto.FavoriteItem, len(favorites))
	for i, f := range favorites {
		p, ok := postMap[f.TargetID]
		if ok {
			items[i] = dto.FavoriteItem{
				ID:        p.ID,
				Type:      uint8(p.Type),
				Title:     p.Title,
				CreatedAt: p.CreatedAt,
			}
		} else {
			// 如果帖子不存在（已删除），可以跳过或返回空
			items[i] = dto.FavoriteItem{ID: f.TargetID, Type: f.TargetType}
		}
	}

	return items, total, nil
}

func (r *PostService) GetDraftService(ctx context.Context, uid uint, page, size int) ([]dto.PostListItem, int64, error) {
	offset := (page - 1) * size

	// 1. 查询总数
	var total int64
	if err := r.postRepo.CountUserDraftPosts(ctx, uid, &total); err != nil {
		return nil, 0, err
	}

	// 2. 查询草稿列表（一次查询就够了）
	var posts []model.Post
	err := r.postRepo.ListUserDraftPosts(ctx, uid, offset, size, posts)

	if err != nil {
		return nil, 0, err
	}

	// 3. 直接转换为返回结构
	items := make([]dto.PostListItem, len(posts))
	for i, p := range posts {
		items[i] = dto.PostListItem{
			ID:       p.ID,
			Type:     uint8(p.Type),
			Title:    p.Title,
			CreateAt: p.CreatedAt,
		}
	}

	return items, total, nil
}
