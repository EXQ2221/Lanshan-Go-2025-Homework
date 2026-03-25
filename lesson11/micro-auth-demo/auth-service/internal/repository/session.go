package repository

import (
	"context"

	"example.com/micro-auth-demo/auth-service/internal/dal/model"
	"gorm.io/gorm"
)

type SessionRepository interface {
	Create(ctx context.Context, session *model.Session) error
	Update(ctx context.Context, session *model.Session) error
	GetBySessionID(ctx context.Context, sessionID string) (*model.Session, error)
	ListByUserID(ctx context.Context, userID int64) ([]model.Session, error)
}

type GormSessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *GormSessionRepository {
	return &GormSessionRepository{db: db}
}

func (r *GormSessionRepository) Create(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *GormSessionRepository) Update(ctx context.Context, session *model.Session) error {
	return r.db.WithContext(ctx).Save(session).Error
}

func (r *GormSessionRepository) GetBySessionID(ctx context.Context, sessionID string) (*model.Session, error) {
	var session model.Session
	if err := r.db.WithContext(ctx).Where("session_id = ?", sessionID).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *GormSessionRepository) ListByUserID(ctx context.Context, userID int64) ([]model.Session, error) {
	var sessions []model.Session
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("last_seen_at desc").Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}
