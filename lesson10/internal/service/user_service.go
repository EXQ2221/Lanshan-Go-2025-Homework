package service

import (
	"context"
	"errors"
	"lesson10/internal/dto"
	"lesson10/internal/model"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/repository"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo   repository.UserRepository
	postRepo   repository.PostRepository
	followRepo repository.FollowRepository
	db         *gorm.DB
	authSvc    *AuthService
}

func NewUserService(userRepo repository.UserRepository, followRepo repository.FollowRepository, postRepo repository.PostRepository, db *gorm.DB) *UserService {
	return &UserService{
		userRepo:   userRepo,
		followRepo: followRepo,
		postRepo:   postRepo,
		db:         db,
	}
}

func (r *UserService) SetAuthService(authSvc *AuthService) {
	r.authSvc = authSvc
}

func (r *UserService) RegisterService(ctx context.Context, req dto.RegisterRequest) (*model.User, error) {
	exists, err := r.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, errcode.ErrInternal
	}
	if exists {
		return nil, errcode.ErrConflict
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	user := &model.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	if err := r.userRepo.CreateUser(ctx, user); err != nil {
		return nil, errcode.ErrInternal
	}

	return user, nil
}

func (r *UserService) LoginService(
	ctx context.Context,
	req dto.LoginRequest,
	ip string,
	ua string,
) (string, string, *model.User, error) {
	if r.authSvc == nil {
		return "", "", nil, errcode.ErrInternal
	}

	pair, user, _, err := r.authSvc.Login(ctx, req, ip, ua)
	if err != nil {
		return "", "", nil, err
	}

	return pair.AccessToken, pair.RefreshToken, user, nil
}

func (r *UserService) ChangePassService(ctx context.Context, req dto.ChangePassRequest, id uint) error {
	var user model.User
	if err := r.userRepo.FindUserByID(ctx, id, &user); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrNotFound
		}
		return errcode.ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPass)); err != nil {
		return errcode.ErrForbidden
	}

	if req.NewPass == "" || req.OldPass == req.NewPass {
		return errcode.ErrForbidden
	}

	newHashBytes, err := bcrypt.GenerateFromPassword([]byte(req.NewPass), bcrypt.DefaultCost)
	if err != nil {
		return errcode.ErrInternal
	}

	if err = r.userRepo.ChangePassWord(ctx, id, string(newHashBytes)); err != nil {
		return errcode.ErrInternal
	}

	if r.authSvc != nil {
		if err := r.authSvc.RevokeAllUserSessions(ctx, id, "password_changed"); err != nil {
			return err
		}
	}

	return nil
}

func (r *UserService) UpdateProfileService(ctx context.Context, req dto.UpdateProfileRequest, id uint) error {
	updates := map[string]any{}

	if req.Profile != nil {
		updates["profile"] = strings.TrimSpace(*req.Profile)
	}

	if len(updates) == 0 {
		return nil
	}

	res := r.userRepo.UpdateUserProfile(ctx, id, updates)
	if res.Error != nil {
		return errcode.ErrInternal
	}
	if res.RowsAffected == 0 {
		return errcode.ErrNotFound
	}

	return nil
}

func (r *UserService) UpdateAvatarService(ctx context.Context, userID uint, avatarURL string) error {
	if err := r.userRepo.UpdateUserAvatar(ctx, userID, avatarURL); err != nil {
		return errcode.ErrInternal
	}

	return nil
}

func (r *UserService) GetUserInfoService(ctx context.Context, currentID, id uint, page int) (*dto.UserPublicInfo, error) {
	var user model.User
	err := r.userRepo.FindUserByID(ctx, id, &user)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrNotFound
	}
	if err != nil {
		return nil, errcode.ErrInternal
	}

	const size = 5
	offset := (page - 1) * size

	posts, err := r.postRepo.ListUserPublicPosts(ctx, id, offset, size)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	total, err := r.postRepo.CountUserPublicPosts(ctx, id)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	if currentID == id {
		draftPosts, err := r.postRepo.ListUserDraftPost(ctx, id, offset, size)
		if err != nil {
			return nil, errcode.ErrInternal
		}

		draftTotal, err := r.postRepo.CountUserDraftPost(ctx, id)
		if err != nil {
			return nil, errcode.ErrInternal
		}

		posts = append(posts, draftPosts...)
		total += draftTotal
	}

	postSummaries := make([]dto.PostSummary, len(posts))
	for i, p := range posts {
		postSummaries[i] = dto.PostSummary{
			ID:        p.ID,
			Title:     p.Title,
			CreatedAt: p.CreatedAt,
			Status:    p.Status,
		}
	}

	isVIP := user.VIPExpiresAt != nil && time.Now().Before(*user.VIPExpiresAt)

	followingCount, err := r.followRepo.CountFollowing(ctx, id)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	followersCount, err := r.followRepo.CountFollowers(ctx, id)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	isFollowed := false
	if currentID > 0 && currentID != id {
		isFollowed, err = r.followRepo.IsFollowing(ctx, currentID, id)
		if err != nil {
			return nil, errcode.ErrInternal
		}
	}

	return &dto.UserPublicInfo{
		ID:             user.ID,
		Username:       user.Username,
		Profile:        user.Profile,
		AvatarURL:      user.AvatarURL,
		Role:           user.Role,
		IsVIP:          isVIP,
		VIPExpiresAt:   user.VIPExpiresAt,
		Posts:          postSummaries,
		PostTotal:      total,
		FollowingCount: followingCount,
		FollowersCount: followersCount,
		IsFollowed:     isFollowed,
		Page:           page,
		Size:           size,
	}, nil
}

func (r *UserService) RefreshWithWhitelist(
	ctx context.Context,
	rawRefresh string,
	userID uint,
	tokenVersion int,
	sid string,
	ip string,
	ua string,
) (string, string, bool, error) {
	if r.authSvc == nil {
		return "", "", false, errcode.ErrInternal
	}

	pair, err := r.authSvc.Refresh(ctx, dto.RefreshRequest{RefreshToken: rawRefresh}, ip, ua)
	if err != nil {
		return "", "", false, err
	}

	return pair.AccessToken, pair.RefreshToken, false, nil
}

func (r *UserService) CreateRefreshSession(
	ctx context.Context,
	userID uint,
	tokenVersion int,
	sid string,
	rawRefresh string,
	ip string,
	ua string,
) error {
	return nil
}
