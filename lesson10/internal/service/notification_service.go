package service

import (
	"context"
	"lesson10/internal/dto"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/repository"
	"log"
)

type NotificationService struct {
	notificationRepo *repository.NotificationRepo
	userRepo         *repository.UserRepo
}

func NewNotificationService(notificationRepo *repository.NotificationRepo, userRepo *repository.UserRepo) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
	}
}

func (r *NotificationService) GetNotifications(ctx context.Context, uid uint, page, size int, unreadOnly bool) ([]dto.NotificationItem, int64, error) {
	offset := (page - 1) * size

	var total int64
	err := r.notificationRepo.CountNotifications(ctx, uid, total)
	if err != nil {
		return nil, 0, err
	}

	notifications, err := r.notificationRepo.ListNotifications(ctx, uid, unreadOnly, offset, size)

	// 批量查触发者用户名（可选）
	actorIDs := make([]uint, 0, len(notifications))
	actorMap, err := r.userRepo.BatchGetUsernames(ctx, actorIDs)
	if err != nil {
		log.Printf("批量查用户名失败: %v", err)
	}

	items := make([]dto.NotificationItem, len(notifications))
	for i, n := range notifications {
		var actorID uint = 0
		if n.ActorID != nil {
			actorID = *n.ActorID
		} else {
			actorID = 0
		}

		if n.ActorID != nil {
			actorID = *n.ActorID
		}

		var targetType uint8 = 0
		if n.TargetType != nil {
			targetType = *n.TargetType
		}

		var targetID uint = 0
		if n.TargetID != nil {
			targetID = *n.TargetID
		}
		items[i] = dto.NotificationItem{
			ID:         n.ID,
			Type:       n.Type,
			ActorID:    actorID,
			ActorName:  actorMap[actorID],
			TargetType: targetType,
			TargetID:   targetID,
			Content:    n.Content,
			IsRead:     n.IsRead == 1,
			CreatedAt:  n.CreatedAt,
		}
	}

	return items, total, nil
}

func (r *NotificationService) GetUnreadCountService(ctx context.Context, uid uint) (int64, error) {
	var count int64
	if err := r.notificationRepo.GetUnreadCount(ctx, uid, count); err != nil {
		return 0, errcode.ErrInternal
	}
	return count, nil
}

func (r *NotificationService) MarkAllNotificationsRead(ctx context.Context, uid uint) error {
	if err := r.notificationRepo.MarkAllNotificationsRead(ctx, uid); err != nil {
		return errcode.ErrInternal
	}
	return nil
}
