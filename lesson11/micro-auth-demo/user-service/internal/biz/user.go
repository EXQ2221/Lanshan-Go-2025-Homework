package biz

import (
	"context"
	"errors"

	"example.com/micro-auth-demo/user-service/internal/dal/model"
	"example.com/micro-auth-demo/user-service/internal/pkg/password"
	"example.com/micro-auth-demo/user-service/internal/repository"
	"gorm.io/gorm"
)

type UserService struct {
	Repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{Repo: repo}
}

func (s *UserService) GetByID(ctx context.Context, id int64) (*model.User, error) {
	return s.Repo.GetByID(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, email, nickname, rawPassword string) (*model.User, error) {
	hash := password.HashPassword(rawPassword)
	if hash == "" {
		return nil, errors.New("failed to hash password")
	}

	user := &model.User{
		Email:        email,
		Nickname:     nickname,
		PasswordHash: hash,
	}

	if err := s.Repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) VerifyCredential(ctx context.Context, email, rawPassword string) (*model.User, bool, error) {
	user, err := s.Repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}

	if !password.ComparePassword(rawPassword, user.PasswordHash) {
		return nil, false, nil
	}

	return user, true, nil
}

func (s *UserService) CheckPassword(ctx context.Context, userID int64, rawPassword string) (bool, error) {
	user, err := s.Repo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return password.ComparePassword(rawPassword, user.PasswordHash), nil
}

func (s *UserService) SeedDemoUser(ctx context.Context) error {
	_, err := s.Repo.GetByEmail(ctx, "demo@example.com")
	switch {
	case err == nil:
		return nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return err
	}

	_, err = s.CreateUser(ctx, "demo@example.com", "demo-user", "Password123!")
	return err
}
