package repository

import (
	"context"
	"lesson10/internal/model"

	"gorm.io/gorm"
)

type FavoriteRepository interface {
	FindFav(ctx context.Context, uid uint, targetType uint8, targetID uint, fav *model.Favorite) error
	DeleteFav(ctx context.Context, uid uint, targetType uint8, targetID uint) *gorm.DB
	CreateFav(ctx context.Context, newFav model.Favorite) error
	CountByUserID(ctx context.Context, userID uint) (int64, error)
	ListByUserID(ctx context.Context, userID uint, offset, limit int, fav *[]model.Favorite) error
}
type favoriteRepo struct {
	db *gorm.DB
}

func NewFavoriteRepo(db *gorm.DB) FavoriteRepository {
	return &favoriteRepo{db: db}
}

func (r *favoriteRepo) FindFav(ctx context.Context, uid uint, targetType uint8, targetID uint, fav *model.Favorite) error {
	err := r.db.WithContext(ctx).Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).
		First(fav).Error
	return err
}

func (r *favoriteRepo) DeleteFav(ctx context.Context, uid uint, targetType uint8, targetID uint) *gorm.DB {
	result := r.db.WithContext(ctx).Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).
		Delete(&model.Favorite{})
	return result
}

func (r *favoriteRepo) CreateFav(ctx context.Context, newFav model.Favorite) error {
	err := r.db.WithContext(ctx).Create(&newFav).Error
	return err
}

func (r *favoriteRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Favorite{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	return total, err
}

func (r *favoriteRepo) ListByUserID(ctx context.Context, userID uint, offset, limit int, fav *[]model.Favorite) error {
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(fav).Error
	return err
}
