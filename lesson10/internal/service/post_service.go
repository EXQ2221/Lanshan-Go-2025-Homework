package service

import (
	"context"
	"errors"
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

func (r *PostService) ListPostsService(ctx context.Context, q dto.ListPostsQuery) ([]dto.PostListItem, int64, error) {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 20
	}
	if q.PageSize > 100 {
		q.PageSize = 100
	}

	items, total, err := r.postRepo.ListPosts(ctx, q)
	if err != nil {
		return nil, 0, errcode.ErrInternal
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
		ID:              p.ID,
		Type:            uint8(p.Type),
		AuthorID:        p.AuthorID,
		AuthorName:      p.Author.Username,
		AuthorAvatarURL: p.Author.AvatarURL,
		Title:           p.Title,
		Content:         p.Content,
		Status:          p.Status,
		LikeCount:       p.LikeCount,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
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
	err := r.postRepo.ListUserDraftPosts(ctx, uid, offset, size, &posts)

	if err != nil {
		return nil, 0, err
	}

	// 3. 直接转换为返回结构
	items := make([]dto.PostListItem, len(posts))
	for i, p := range posts {
		items[i] = dto.PostListItem{
			ID:              p.ID,
			Type:            uint8(p.Type),
			AuthorID:        p.AuthorID,
			AuthorName:      p.Author.Username,
			AuthorAvatarURL: p.Author.AvatarURL,
			Title:           p.Title,
			CreatedAt:       p.CreatedAt,
		}
	}

	return items, total, nil
}
