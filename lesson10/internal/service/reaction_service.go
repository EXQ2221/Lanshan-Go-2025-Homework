package service

import (
	"context"
	"errors"
	"lesson10/internal/model"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/repository"

	"gorm.io/gorm"
)

type ReactionService struct {
	reactionRepo     repository.ReactionRepository
	postRepo         repository.PostRepository
	commentRepo      repository.CommentRepository
	notificationRepo repository.NotificationRepository
	db               *gorm.DB
}

func NewReactionService(reactionRepo repository.ReactionRepository, postRepo repository.PostRepository, commentRepo repository.CommentRepository, notificationRepo repository.NotificationRepository, db *gorm.DB) *ReactionService {
	return &ReactionService{
		reactionRepo:     reactionRepo,
		postRepo:         postRepo,
		commentRepo:      commentRepo,
		notificationRepo: notificationRepo,
		db:               db,
	}
}

// ToggleReactionService 切换点赞状态，返回操作后的“是否已点赞”
func (r *ReactionService) ToggleReactionService(ctx context.Context, uid uint, targetType uint8, targetID uint) (*bool, error) {
	   // 事务保证点赞状态和 like_count 一致性
	   var isLiked *bool
	   err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 检查目标合法性
		if targetType == 1 || targetType == 2 {
			var post model.Post
			err := r.postRepo.FindPostByID(ctx, targetID, &post)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.ErrNotFound
			}
			if err != nil {
				return errcode.ErrInternal
			}
			if post.Status == 1 && post.AuthorID != uid {
				return errcode.ErrUnauthorized
			}
		}
		// 检查是否已点赞
		var reaction model.Reaction
		err := tx.Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).First(&reaction).Error
		if err == nil {
			// 已点赞 → 取消
			if err := tx.Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).Delete(&model.Reaction{}).Error; err != nil {
				return errcode.ErrInternal
			}
			// like_count -1
			switch targetType {
			case 1, 2:
				if err := tx.Model(&model.Post{}).Where("id = ?", targetID).Update("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
					return errcode.ErrInternal
				}
			case 3:
				if err := tx.Model(&model.Comment{}).Where("id = ?", targetID).Update("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
					return errcode.ErrInternal
				}
			}
			b := false
			isLiked = &b
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrInternal
		}
		// 未点赞 → 添加
		newReaction := &model.Reaction{
			UserID:     uid,
			TargetType: targetType,
			TargetID:   targetID,
		}
		if err := tx.Create(newReaction).Error; err != nil {
			return errcode.ErrInternal
		}
		// like_count +1
		switch targetType {
		case 1, 2:
			if err := tx.Model(&model.Post{}).Where("id = ?", targetID).Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
				return errcode.ErrInternal
			}
		case 3:
			if err := tx.Model(&model.Comment{}).Where("id = ?", targetID).Update("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
				return errcode.ErrInternal
			}
		}
		// 通知
		if targetType == 1 || targetType == 2 || targetType == 3 {
			var receiverID uint
			if targetType == 1 || targetType == 2 {
				var post model.Post
				if err := r.postRepo.FindPostByID(ctx, targetID, &post); err == nil {
					receiverID = post.AuthorID
				}
			} else {
				var comment model.Comment
				if err := r.commentRepo.GetAuthorIDByComment(ctx, targetID, &comment); err == nil {
					receiverID = comment.AuthorID
				}
			}
			if receiverID != 0 && receiverID != uid {
				notifyType := uint8(2)
				content := "有人点赞了你的内容"
				tt := targetType
				tid := targetID
				notification := &model.Notification{
					UserID:     receiverID,
					Type:       notifyType,
					ActorID:    &uid,
					TargetType: &tt,
					TargetID:   &tid,
					Content:    content,
				}
				_ = r.notificationRepo.CreateNotification(ctx, notification)
			}
		}
		b := true
		isLiked = &b
		return nil
	})
	if err != nil {
		return nil, err
	}
	return isLiked, nil
}
