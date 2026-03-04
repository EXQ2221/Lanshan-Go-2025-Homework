package repository

import (
	"context"
	"lesson10/internal/dto"
	"lesson10/internal/model"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("username = ?", username).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepo) ExistsByUserID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Count(&count).Error
	return count > 0, err
}

func (r *UserRepo) FindUserByID(ctx context.Context, id uint, user *model.User) error {
	err := r.db.WithContext(ctx).Where("id = ?", id).First(user).Error
	return err
}

func (r *UserRepo) ChangePassWord(ctx context.Context, id uint, newHash string) error {
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"password_hash": newHash,
			"token_version": gorm.Expr("token_version + 1"),
		}).Error
	return err
}

func (r *UserRepo) UpdateUserProfile(ctx context.Context, id uint, updates map[string]any) *gorm.DB {

	res := r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", id).
		Updates(updates)
	return res
}

func (r *UserRepo) UpdateUserAvatar(ctx context.Context, userID uint, avatarURL string) error {
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("avatar_url", avatarURL).Error
	return err
}

func (r *UserRepo) CreateUser(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *UserRepo) BatchGetAuthorUsernames(ctx context.Context, authorIDs []uint) (map[uint]string, error) {
	if len(authorIDs) == 0 {
		return make(map[uint]string), nil
	}

	var authors []struct {
		ID       uint
		Username string
	}

	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Select("id, username").
		Where("id IN ?", authorIDs).
		Find(&authors).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]string, len(authors))
	for _, a := range authors {
		result[a.ID] = a.Username
	}
	return result, nil
}

func (r *UserRepo) BatchGetUserBasicInfo(ctx context.Context, userIDs []uint) (map[uint]dto.UserBasicInfo, error) {
	if len(userIDs) == 0 {
		return make(map[uint]dto.UserBasicInfo), nil
	}

	var users []struct {
		ID        uint
		Username  string
		AvatarURL string
		Profile   string
	}

	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Select("id, username, avatar_url, profile").
		Where("id IN ?", userIDs).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]dto.UserBasicInfo, len(users))
	for _, u := range users {
		result[u.ID] = dto.UserBasicInfo{
			Username:  u.Username,
			AvatarURL: u.AvatarURL,
			Profile:   u.Profile,
		}
	}

	return result, nil
}

func (r *UserRepo) BatchGetUsernames(ctx context.Context, userIDs []uint) (map[uint]string, error) {
	if len(userIDs) == 0 {
		return make(map[uint]string), nil
	}

	var users []struct {
		ID       uint
		Username string
	}

	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Select("id, username").
		Where("id IN ?", userIDs).
		Find(&users).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uint]string, len(users))
	for _, u := range users {
		result[u.ID] = u.Username
	}
	return result, nil
}

func (r *UserRepo) RefreshTokenVersion(ctx context.Context, userID uint, currentVersion int) (bool, error) {
	res := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ? AND token_version = ?", userID, currentVersion).
		Update("token_version", gorm.Expr("token_version + 1"))

	if res.Error != nil {
		return false, res.Error
	}

	return res.RowsAffected > 0, nil
}

func (r *UserRepo) FindUserForToken(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Select("id", "username", "role", "token_version").
		Where("id = ?", userID).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
