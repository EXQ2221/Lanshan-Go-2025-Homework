package repository

import (
	"context"
	"lesson10/internal/model"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification *model.Notification) error
	GetUnreadCount(ctx context.Context, uid uint, count int64) error
	ListNotifications(ctx context.Context, userID uint, unreadOnly bool, offset, limit int) ([]model.Notification, error)
	MarkAllNotificationsRead(ctx context.Context, uid uint) error
	CountNotifications(ctx context.Context, uid uint, total *int64) error
	// 后续加方法就直接在这里加
}
type notificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) NotificationRepository {
	return &notificationRepo{db: db}
}

func (r *notificationRepo) CreateNotification(ctx context.Context, notification *model.Notification) error {
	err := r.db.WithContext(ctx).Create(notification).Error
	return err
}

func (r *notificationRepo) CountNotifications(ctx context.Context, uid uint, total *int64) error {
	err := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ?", uid).
		Count(total).Error
	return err
}

func (r *notificationRepo) ListNotifications(ctx context.Context, userID uint, unreadOnly bool, offset, limit int) ([]model.Notification, error) {
	var notifications []model.Notification

	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit)

	if unreadOnly {
		query = query.Where("is_read = 0")
	}

	err := query.Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepo) GetUnreadCount(ctx context.Context, uid uint, count int64) error {
	err := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = 0", uid).
		Count(&count).Error
	return err
}

func (r *notificationRepo) MarkAllNotificationsRead(ctx context.Context, uid uint) error {
	err := r.db.WithContext(ctx).Model(&model.Notification{}).
		Where("user_id = ? AND is_read = 0", uid).
		Update("is_read", 1).Error
	return err
}
