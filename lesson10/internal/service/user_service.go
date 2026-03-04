package service

import (
	"context"
	"errors"
	"lesson10/internal/dto"
	"lesson10/internal/model"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/pkg/token"
	"lesson10/internal/repository"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo   *repository.UserRepo
	postRepo   *repository.PostRepo
	followRepo *repository.FollowRepo
}

func NewUserService(userRepo *repository.UserRepo, followRepo *repository.FollowRepo, postRepo *repository.PostRepo) *UserService {
	return &UserService{
		userRepo:   userRepo,
		followRepo: followRepo,
		postRepo:   postRepo,
	}
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

	err = r.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, errcode.ErrInternal
	}
	return user, nil
}

func (r *UserService) LoginService(ctx context.Context, req dto.LoginRequest) (string, *model.User, error) {
	// 1) 按用户名查用户
	user, err := r.userRepo.FindUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil, errors.New("username incorrect")
		}
		return "", nil, errcode.ErrInternal
	}

	// 2) 校验密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Println("password incorrect:", err)
		return "", nil, errors.New("password incorrect")
	}

	// 3) 生成 access token
	tokenRes, err := token.GenerateToken(user.Username, user.ID, user.TokenVersion, user.Role)
	if err != nil {
		log.Println("login error tk:", err)
		return "", user, errcode.ErrInternal
	}

	return tokenRes, user, nil
}
func (r *UserService) ChangePassService(ctx context.Context, req dto.ChangePassRequest, id uint) error {
	var user model.User
	exists, err := r.userRepo.ExistsByUserID(ctx, id)
	if err != nil {
		return errcode.ErrInternal
	}
	if exists {
		return errcode.ErrConflict
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

	newHash := string(newHashBytes)

	if err = r.userRepo.ChangePassWord(ctx, id, newHash); err != nil {
		return errcode.ErrInternal
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

	var posts []model.Post
	var total int64
	var size = 5
	offset := (page - 1) * size

	publicPosts, err := r.postRepo.ListUserPublicPosts(ctx, id, offset, size)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	publicTotal, err := r.postRepo.CountUserPublicPosts(ctx, id)
	if err != nil {
		return nil, errcode.ErrInternal
	}

	total = publicTotal
	posts = publicPosts
	// 如果当前用户是本人，额外查草稿
	var draftPosts []model.Post
	var draftTotal int64
	if currentID == id {
		draftPosts, err = r.postRepo.ListUserDraftPost(ctx, id, offset, size)
		if err != nil {
			return nil, errcode.ErrInternal
		}

		draftTotal, err = r.postRepo.CountUserDraftPost(ctx, id)
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

	userPublicInfo := dto.UserPublicInfo{
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
	}

	return &userPublicInfo, nil
}

func (r *UserService) Refresh(ctx context.Context, userID uint, tokenVersion int) (string, string, error) {
	// 1. 查用户基础信息（用户名、角色、当前版本号）
	user, err := r.userRepo.FindUserForToken(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errcode.ErrUnauthorized
		}
		return "", "", errcode.ErrInternal
	}

	// 2. 校验版本号
	if user.TokenVersion != tokenVersion {
		return "", "", errcode.ErrUnauthorized
	}

	// 3. 更新版本号（原子操作）
	updated, err := r.userRepo.RefreshTokenVersion(ctx, userID, tokenVersion)
	if err != nil {
		log.Printf("更新 token 版本失败: %v", err)
		return "", "", errcode.ErrInternal
	}
	if !updated {
		return "", "", errcode.ErrUnauthorized // 并发更新失败，版本已变
	}

	// 4. 生成新 token（版本号已 +1）
	newVersion := tokenVersion + 1
	accessToken, err := token.GenerateToken(user.Username, user.ID, newVersion, user.Role)
	if err != nil {
		return "", "", errcode.ErrInternal
	}

	refreshToken, err := token.GenerateRefreshToken(user.ID, newVersion)
	if err != nil {
		return "", "", errcode.ErrInternal
	}

	return accessToken, refreshToken, nil
}
