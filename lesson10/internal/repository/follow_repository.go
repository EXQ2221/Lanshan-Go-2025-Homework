package repository

import (
	"context"
	"lesson10/internal/model"

	"gorm.io/gorm"
)

type FollowRepository interface {
	CountFollowing(ctx context.Context, userID uint) (int64, error)
	CountFollowers(ctx context.Context, userID uint) (int64, error)
	IsFollowing(ctx context.Context, followerID, followeeID uint) (bool, error)
	CreateFollow(ctx context.Context, follow model.UserFollow) error
	DeleteFollow(ctx context.Context, followerID, followeeID uint) *gorm.DB
	CountFollows(ctx context.Context, targetUserID uint, isFollowers bool) (int64, error)
	ListFollowIDs(ctx context.Context, targetUserID uint, isFollowers bool, offset, limit int) ([]uint, error)
	BatchIsFollowing(ctx context.Context, currentUserID uint, targetUserIDs []uint) (map[uint]bool, error)
}
type followRepo struct {
	db *gorm.DB
}

func NewFollowRepo(db *gorm.DB) FollowRepository {
	return &followRepo{db: db}
}

func (r *followRepo) CountFollowing(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserFollow{}).
		Where("follower_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *followRepo) CountFollowers(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserFollow{}).
		Where("followee_id = ?", userID).
		Count(&count).Error
	return count, err
}

func (r *followRepo) IsFollowing(ctx context.Context, followerID, followeeID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserFollow{}).
		Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Count(&count).Error
	return count > 0, err
}

func (r *followRepo) CreateFollow(ctx context.Context, follow model.UserFollow) error {
	err := r.db.WithContext(ctx).Create(&follow).Error
	return err
}

func (r *followRepo) DeleteFollow(ctx context.Context, followerID, followeeID uint) *gorm.DB {
	result := r.db.WithContext(ctx).Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Delete(&model.UserFollow{})
	return result
}

func (r *followRepo) CountFollows(ctx context.Context, targetUserID uint, isFollowers bool) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&model.UserFollow{})

	if isFollowers {
		query = query.Where("followee_id = ?", targetUserID)
	} else {
		query = query.Where("follower_id = ?", targetUserID)
	}

	err := query.Count(&count).Error
	return count, err
}

func (r *followRepo) ListFollowIDs(ctx context.Context, targetUserID uint, isFollowers bool, offset, limit int) ([]uint, error) {
	var ids []uint
	query := r.db.WithContext(ctx).Model(&model.UserFollow{})

	if isFollowers {
		query = query.Where("followee_id = ?", targetUserID).Select("follower_id")
	} else {
		query = query.Where("follower_id = ?", targetUserID).Select("followee_id")
	}

	err := query.Offset(offset).Limit(limit).Pluck("id", &ids).Error
	return ids, err
}

func (r *followRepo) BatchIsFollowing(ctx context.Context, currentUserID uint, targetUserIDs []uint) (map[uint]bool, error) {
	if currentUserID == 0 || len(targetUserIDs) == 0 {
		return make(map[uint]bool), nil
	}

	var follows []model.UserFollow
	err := r.db.WithContext(ctx).
		Where("follower_id = ? AND followee_id IN ?", currentUserID, targetUserIDs).
		Find(&follows).Error
	if err != nil {
		return nil, err
	}

	likedMap := make(map[uint]bool, len(follows))
	for _, f := range follows {
		likedMap[f.FolloweeID] = true
	}
	return likedMap, nil
}
