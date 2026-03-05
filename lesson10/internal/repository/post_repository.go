package repository

import (
	"context"
	"lesson10/internal/model"

	"gorm.io/gorm"
)

type PostRepository interface {
	CreatePost(ctx context.Context, p *model.Post) error
	FindPostByID(ctx context.Context, id uint, p *model.Post) error
	ExistsPostByID(ctx context.Context, id uint) (bool, error)
	UpdatePost(ctx context.Context, updates map[string]interface{}, post model.Post) error
	DeletePost(ctx context.Context, post model.Post) error
	ListUserPublicPosts(ctx context.Context, userID uint, offset, limit int) ([]model.Post, error)
	CountUserPublicPosts(ctx context.Context, userID uint) (int64, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)
	GetAuthorIDByPost(ctx context.Context, targetID uint, post *model.Post) error
	FindPostsByIDs(ctx context.Context, postIDs []uint, posts *[]model.Post)
	ListUserDraftPost(ctx context.Context, userID uint, offset, limit int) ([]model.Post, error)
	CountUserDraftPost(ctx context.Context, userID uint) (int64, error)
	ListUserDraftPosts(ctx context.Context, userID uint, offset, limit int, posts []model.Post) error
	CountUserDraftPosts(ctx context.Context, uid uint, total *int64) error
}

type postRepo struct {
	db *gorm.DB
}

func NewPostRepo(db *gorm.DB) PostRepository {
	return &postRepo{db: db}
}

func (r *postRepo) CreatePost(ctx context.Context, p *model.Post) error {
	err := r.db.WithContext(ctx).Create(p).Error
	return err
}

func (r *postRepo) FindPostByID(ctx context.Context, id uint, p *model.Post) error {
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_deleted = 0", id).
		Preload("Author").First(p).Error

	return err
}

func (r *postRepo) ExistsPostByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Post{}).
		Where("id = ? AND is_deleted = 0", id).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *postRepo) UpdatePost(ctx context.Context, updates map[string]interface{}, post model.Post) error {
	err := r.db.WithContext(ctx).Model(&post).Updates(updates).Error
	return err
}

func (r *postRepo) DeletePost(ctx context.Context, post model.Post) error {
	err := r.db.WithContext(ctx).Model(&post).Update("is_deleted", 1).Error
	return err
}

func (r *postRepo) ListUserPublicPosts(ctx context.Context, userID uint, offset, limit int) ([]model.Post, error) {
	var posts []model.Post
	err := r.db.WithContext(ctx).
		Where("author_id = ? AND is_deleted = 0 AND status = 0", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

func (r *postRepo) CountUserPublicPosts(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Post{}).
		Where("author_id = ? AND is_deleted = 0 AND status = 0", userID).
		Count(&total).Error
	return total, err
}

func (r *postRepo) ListUserDraftPosts(ctx context.Context, userID uint, offset, limit int, posts []model.Post) error {
	err := r.db.WithContext(ctx).
		Where("author_id = ? AND is_deleted = 0 AND status = 1", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error
	return err
}

func (r *postRepo) CountUserDraftPosts(ctx context.Context, uid uint, total *int64) error {

	err := r.db.WithContext(ctx).
		Model(&model.Post{}).
		Where("author_id = ? AND status = 1 AND is_deleted = 0", uid).
		Count(total).Error
	return err
}

func (r *postRepo) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Post{}).
		Where("id = ? AND is_deleted = 0", id).
		Count(&count).Error
	return count > 0, err
}

func (r *postRepo) GetAuthorIDByPost(ctx context.Context, targetID uint, post *model.Post) error {
	err := r.db.WithContext(ctx).
		Select("author_id").Where("id = ?", targetID).First(post).Error
	return err
}

func (r *postRepo) FindPostsByIDs(ctx context.Context, postIDs []uint, posts *[]model.Post) {
	r.db.WithContext(ctx).Select("id, type, title, created_at").
		Where("id IN ? AND is_deleted = 0", postIDs).
		Find(posts)
}

func (r *postRepo) ListUserDraftPost(ctx context.Context, userID uint, offset, limit int) ([]model.Post, error) {
	var posts []model.Post
	err := r.db.WithContext(ctx).
		Where("author_id = ? AND is_deleted = 0 AND status = 1", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error
	return posts, err
}

func (r *postRepo) CountUserDraftPost(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Post{}).
		Where("author_id = ? AND is_deleted = 0 AND status = 1", userID).
		Count(&total).Error
	return total, err
}
