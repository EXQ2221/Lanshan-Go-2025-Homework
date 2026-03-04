package repository

import (
	"context"
	"lesson10/internal/model"

	"gorm.io/gorm"
)

type FavoriteRepo struct {
	db *gorm.DB
}

func NewFavoriteRepo(db *gorm.DB) *FavoriteRepo {
	return &FavoriteRepo{db: db}
}

func (r *FavoriteRepo) FindFav(ctx context.Context, uid uint, targetType uint8, targetID uint, fav *model.Favorite) error {
	err := r.db.WithContext(ctx).Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).
		First(fav).Error
	return err
}

func (r *FavoriteRepo) DeleteFav(ctx context.Context, uid uint, targetType uint8, targetID uint) *gorm.DB {
	result := r.db.WithContext(ctx).Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).
		Delete(&model.Favorite{})
	return result
}

func (r *FavoriteRepo) CreateFav(ctx context.Context, newFav model.Favorite) error {
	err := r.db.WithContext(ctx).Create(&newFav).Error
	return err
}

func (r *FavoriteRepo) CountByUserID(ctx context.Context, userID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Favorite{}).
		Where("user_id = ?", userID).
		Count(&total).Error
	return total, err
}

func (r *FavoriteRepo) ListByUserID(ctx context.Context, userID uint, offset, limit int, fav *[]model.Favorite) error {
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Offset(offset).
		Limit(limit).
		Find(fav).Error
	return err
}
