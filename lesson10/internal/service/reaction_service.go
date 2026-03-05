package service

import (
	"context"
	"errors"
	"lesson10/internal/model"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/repository"
	"log"
	"time"

	"gorm.io/gorm"
)

type ReactionService struct {
	reactionRepo     repository.ReactionRepository
	postRepo         repository.PostRepository
	commentRepo      repository.CommentRepository
	notificationRepo repository.NotificationRepository
}

func NewReactionService(reactionRepo repository.ReactionRepository, postRepo repository.PostRepository, commentRepo repository.CommentRepository, notificationRepo repository.NotificationRepository) *ReactionService {
	return &ReactionService{
		reactionRepo:     reactionRepo,
		postRepo:         postRepo,
		commentRepo:      commentRepo,
		notificationRepo: notificationRepo,
	}
}

// ToggleReactionService 切换点赞状态，返回操作后的“是否已点赞”
func (r *ReactionService) ToggleReactionService(ctx context.Context, uid uint, targetType uint8, targetID uint) (*bool, error) {

	const maxRetries = 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(50 * time.Millisecond) // 等待前操作完成
		}

		if targetType == 1 || targetType == 2 {
			var post model.Post
			err := r.postRepo.FindPostByID(ctx, targetID, &post)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errcode.ErrNotFound
			}
			if err != nil {
				return nil, errcode.ErrInternal
			}

			// 草稿状态（status = 1）只能作者本人点赞
			if post.Status == 1 && post.AuthorID != uid {
				return nil, errcode.ErrUnauthorized
			}
		}
		// 检查是否已点赞
		var reaction model.Reaction
		err := r.reactionRepo.FindReaction(ctx, uid, uint(targetType), targetID, &reaction)

		if err == nil {
			// 已点赞 → 取消（直接用唯一条件删除）
			result := r.reactionRepo.DeleteByUserAndTarget(ctx, uid, targetType, targetID)

			if result.Error != nil {
				log.Printf("delete failed (attempt %d): %v", attempt, result.Error)
				continue
			}

			log.Printf("delete rows affected: %d", result.RowsAffected)

			r.reactionRepo.DecrementLikeCount(ctx, targetType, targetID)
			isLiked := false
			return &isLiked, nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("first reaction failed (attempt %d): %v", attempt, err)
			continue
		}

		// 未点赞 → 添加
		newReaction := &model.Reaction{
			UserID:     uid,
			TargetType: targetType,
			TargetID:   targetID,
		}

		if err = r.reactionRepo.CreateReaction(ctx, newReaction); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				log.Printf("duplicate key on create (attempt %d), retrying...", attempt)
				continue
			}
			log.Printf("create failed (attempt %d): %v", attempt, err)
			return nil, errcode.ErrInternal
		}

		if targetType == 1 || targetType == 2 || targetType == 3 {
			var receiverID uint

			if targetType == 1 || targetType == 2 {
				var post model.Post
				if err = r.postRepo.GetAuthorIDByPost(ctx, targetID, &post); err == nil {
					receiverID = post.AuthorID
				}
			} else {
				var comment model.Comment
				if err = r.commentRepo.GetAuthorIDByComment(ctx, targetID, &comment); err == nil {
					receiverID = comment.AuthorID
				}
			}

			if receiverID != 0 && receiverID != uid {
				notifyType := uint8(2) // 2表示点赞通知
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

		r.reactionRepo.IncrementLikeCount(ctx, targetType, targetID)
		isLiked := true

		return &isLiked, nil
	}

	return nil, errcode.ErrInternal // 重试失败
}
