package repository

import (
	"context"
	"lesson10/internal/model"

	"gorm.io/gorm"
)

type ReactionRepository interface {
	BatchCheckLikedByUser(ctx context.Context, userID uint, commentIDs []uint) (map[uint]bool, error)
	FindReaction(ctx context.Context, uid uint, targetType uint, targetID uint, reaction *model.Reaction) error
	ExistsByUserAndTarget(ctx context.Context, userID uint, targetType uint8, targetID uint) (bool, error)
	CreateReaction(ctx context.Context, reaction *model.Reaction) error
	DeleteByUserAndTarget(ctx context.Context, userID uint, targetType uint8, targetID uint) *gorm.DB
	IncrementLikeCount(ctx context.Context, targetType uint8, targetID uint)
	DecrementLikeCount(ctx context.Context, targetType uint8, targetID uint)
}
type reactionRepo struct {
	db *gorm.DB
}

func NewReactionRepo(db *gorm.DB) ReactionRepository {
	return &reactionRepo{db: db}
}

func (r *reactionRepo) BatchCheckLikedByUser(ctx context.Context, userID uint, commentIDs []uint) (map[uint]bool, error) {
	if userID == 0 || len(commentIDs) == 0 {
		return make(map[uint]bool), nil
	}

	var reactions []model.Reaction
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND target_type = 3 AND target_id IN ? AND is_deleted = 0", userID, commentIDs).
		Find(&reactions).Error
	if err != nil {
		return nil, err
	}

	likedMap := make(map[uint]bool, len(reactions))
	for _, reaction := range reactions {
		likedMap[reaction.TargetID] = true
	}

	return likedMap, nil
}

func (r *reactionRepo) FindReaction(ctx context.Context, uid uint, targetType uint, targetID uint, reaction *model.Reaction) error {
	err := r.db.WithContext(ctx).Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).
		First(reaction).Error
	return err
}

func (r *reactionRepo) ExistsByUserAndTarget(ctx context.Context, userID uint, targetType uint8, targetID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Reaction{}).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Count(&count).Error
	return count > 0, err
}

func (r *reactionRepo) CreateReaction(ctx context.Context, reaction *model.Reaction) error {
	return r.db.WithContext(ctx).Create(reaction).Error
}

func (r *reactionRepo) DeleteByUserAndTarget(ctx context.Context, userID uint, targetType uint8, targetID uint) *gorm.DB {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND target_type = ? AND target_id = ?", userID, targetType, targetID).
		Delete(&model.Reaction{})
	return result
}

func (r *reactionRepo) IncrementLikeCount(ctx context.Context, targetType uint8, targetID uint) {
	switch targetType {
	case 1, 2:
		r.db.WithContext(ctx).Model(&model.Post{}).
			Where("id = ?", targetID).
			Update("like_count", gorm.Expr("like_count + 1"))
	case 3:
		r.db.WithContext(ctx).Model(&model.Comment{}).
			Where("id = ?", targetID).
			Update("like_count", gorm.Expr("like_count + 1"))
	}
}

func (r *reactionRepo) DecrementLikeCount(ctx context.Context, targetType uint8, targetID uint) {
	switch targetType {
	case 1, 2:
		r.db.WithContext(ctx).Model(&model.Post{}).
			Where("id = ?", targetID).
			Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
	case 3:
		r.db.WithContext(ctx).Model(&model.Comment{}).
			Where("id = ?", targetID).
			Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
	}
}
