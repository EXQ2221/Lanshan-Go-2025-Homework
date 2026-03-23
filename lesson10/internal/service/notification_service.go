package service

import (
	"context"
	"lesson10/internal/dto"
	"lesson10/internal/model"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/repository"
	"log"
)

type NotificationService struct {
	notificationRepo repository.NotificationRepository
	userRepo         repository.UserRepository
}

func NewNotificationService(notificationRepo repository.NotificationRepository, userRepo repository.UserRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
	}
}

func (r *NotificationService) GetNotifications(ctx context.Context, uid uint, page, size int, unreadOnly bool) ([]dto.NotificationItem, int64, error) {
	offset := (page - 1) * size

	var total int64
	if err := r.notificationRepo.CountNotifications(ctx, uid, unreadOnly, &total); err != nil {
		return nil, 0, err
	}

	notifications, err := r.notificationRepo.ListNotifications(ctx, uid, unreadOnly, offset, size)
	if err != nil {
		return nil, 0, err
	}

	actorIDs := make([]uint, 0, len(notifications))
	seenActorIDs := make(map[uint]struct{}, len(notifications))
	for _, n := range notifications {
		if n.ActorID == nil || *n.ActorID == 0 {
			continue
		}

		actorID := *n.ActorID
		if _, ok := seenActorIDs[actorID]; ok {
			continue
		}

		seenActorIDs[actorID] = struct{}{}
		actorIDs = append(actorIDs, actorID)
	}

	actorMap, err := r.userRepo.BatchGetUsernames(ctx, actorIDs)
	if err != nil {
		log.Printf("batch get usernames failed: %v", err)
	}

	for _, actorID := range actorIDs {
		if actorMap[actorID] != "" {
			continue
		}

		var actor model.User
		if err := r.userRepo.FindUserByID(ctx, actorID, &actor); err != nil {
			continue
		}

		actorMap[actorID] = actor.Username
	}

	items := make([]dto.NotificationItem, len(notifications))
	for i, n := range notifications {
		var actorID uint
		if n.ActorID != nil {
			actorID = *n.ActorID
		}

		var targetType uint8
		if n.TargetType != nil {
			targetType = *n.TargetType
		}

		var targetID uint
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
	if uid == 0 {
		return 0, errcode.ErrUnauthorized
	}

	var count int64
	if err := r.notificationRepo.GetUnreadCount(ctx, uid, &count); err != nil {
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
