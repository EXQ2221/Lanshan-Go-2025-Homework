package service

import (
	"context"
	"errors"
	"lesson10/internal/dto"
	"lesson10/internal/model"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/pkg/token"
	"lesson10/internal/pkg/utils"
	"lesson10/internal/repository"
	"log"
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
}

func NewUserService(userRepo repository.UserRepository, followRepo repository.FollowRepository, postRepo repository.PostRepository, db *gorm.DB) *UserService {
	return &UserService{
		userRepo:   userRepo,
		followRepo: followRepo,
		postRepo:   postRepo,
		db:         db,
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

func (r *UserService) LoginService(
	ctx context.Context,
	req dto.LoginRequest,
	ip string,
	ua string,
) (string, string, *model.User, error) {
	user, err := r.userRepo.FindUserByUsername(ctx, req.Username)
	   if err != nil {
		   if errors.Is(err, gorm.ErrRecordNotFound) {
			   return "", "", nil, errcode.ErrUsernameIncorrect
		   }
		   return "", "", nil, errcode.ErrInternal
	   }

	   if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		   log.Println("password incorrect:", err)
		   return "", "", nil, errcode.ErrPasswordIncorrect
	   }

	accessToken, err := token.GenerateToken(user.Username, user.ID, user.TokenVersion, user.Role)
	if err != nil {
		log.Println("login error access token:", err)
		return "", "", nil, errcode.ErrInternal
	}

	sid, err := utils.NewSID()
	if err != nil {
		log.Println("login error sid:", err)
		return "", "", nil, errcode.ErrInternal
	}

	refreshToken, err := token.GenerateRefreshToken(user.ID, user.TokenVersion, sid)
	if err != nil {
		log.Println("login error refresh token:", err)
		return "", "", nil, errcode.ErrInternal
	}

	now := time.Now()
	session := &model.RefreshSession{
		SID:              sid,
		UserID:           user.ID,
		TokenVersion:     user.TokenVersion,
		RefreshTokenHash: utils.HashToken(refreshToken),
		ExpiresAt:        time.Now().Add(7 * 24 * time.Hour),
	}

	if ip != "" {
		session.CreatedIP = &ip
	}
	if ua != "" {
		session.CreatedUA = &ua
	}

	if err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.userRepo.RevokeAllActiveRefreshSessions(ctx, tx, user.ID, now); err != nil {
			return err
		}
		if err := r.userRepo.CreateRefreshSession(ctx, tx, session); err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Printf("login rotate sessions failed: user_id=%d err=%v", user.ID, err)
		return "", "", nil, errcode.ErrInternal
	}

	log.Printf("login create session success: sid=%s", sid)
	return accessToken, refreshToken, user, nil
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

func (r *UserService) RefreshWithWhitelist(
	ctx context.Context,
	rawRefresh string,
	userID uint,
	tokenVersion int,
	sid string,
	ip string,
	ua string,
) (string, string, bool, error) {

	var newAccess string
	var newRefresh string
	needRelogin := false

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) 查白名单会话并加锁
		session, err := r.userRepo.GetBySIDForUpdate(ctx, tx, sid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errcode.ErrUnauthorized
			}
			return errcode.ErrInternal
		}

		// 2) 白名单校验
		now := time.Now()
		if session.RevokedAt != nil || session.ExpiresAt.Before(now) {
			return errcode.ErrUnauthorized
		}
		if session.UserID != userID || session.TokenVersion != tokenVersion {
			return errcode.ErrUnauthorized
		}
		if session.RefreshTokenHash != utils.HashToken(rawRefresh) {
			return errcode.ErrUnauthorized
		}

		// 新增：强校验 IP/UA，若变更则允许刷新但标记 needRelogin
		if session.CreatedIP != nil && *session.CreatedIP != ip {
			needRelogin = true
		}
		if session.CreatedUA != nil && *session.CreatedUA != ua {
			needRelogin = true
		}

		// 3) 原子 +1 token_version
		ok, err := r.userRepo.RefreshTokenVersionTx(ctx, tx, userID, tokenVersion)
		if err != nil {
			return errcode.ErrInternal
		}
		if !ok {
			return errcode.ErrUnauthorized
		}

		// 4) 读取用户基本信息（发新 access 需要 username/role）
		user, err := r.userRepo.FindUserForTokenTx(ctx, tx, userID)
		if err != nil {
			return errcode.ErrInternal
		}

		newVersion := tokenVersion + 1
		newSID, err := utils.NewSID()

		// 5) 生成新 token（refresh 带 sid）
		newAccess, err = token.GenerateToken(user.Username, user.ID, newVersion, user.Role)
		if err != nil {
			return errcode.ErrInternal
		}
		newRefresh, err = token.GenerateRefreshToken(user.ID, newVersion, newSID) // 你在 token 包补这个函数
		if err != nil {
			return errcode.ErrInternal
		}

		// 6) 旧会话撤销 + 新会话入白名单
		if err = r.userRepo.RevokeAndReplace(ctx, tx, sid, newSID, now); err != nil {
			return errcode.ErrInternal
		}

		expiresAt := now.Add(7 * 24 * time.Hour) // 按你项目配置改
		newSession := &model.RefreshSession{
			SID:              newSID,
			UserID:           userID,
			TokenVersion:     newVersion,
			RefreshTokenHash: utils.HashToken(newRefresh),
			ExpiresAt:        expiresAt,
			CreatedIP:        &ip,
			CreatedUA:        &ua,
		}
		if err = r.userRepo.CreateRefreshSession(ctx, tx, newSession); err != nil {
			return errcode.ErrInternal
		}

		return nil
	})

	if err != nil {
		return "", "", false, err
	}
	return newAccess, newRefresh, needRelogin, nil
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
	if sid == "" || rawRefresh == "" {
		return errcode.ErrBadRequest
	}
	if r.db == nil {
		return errcode.ErrInternal
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	session := &model.RefreshSession{
		SID:              sid,
		UserID:           userID,
		TokenVersion:     tokenVersion,
		RefreshTokenHash: utils.HashToken(rawRefresh),
		ExpiresAt:        expiresAt,
	}

	if ip != "" {
		session.CreatedIP = &ip
	}
	if ua != "" {
		session.CreatedUA = &ua
	}

	if err := r.userRepo.CreateRefreshSession(ctx, r.db, session); err != nil {
		log.Printf("create refresh session failed: user_id=%d sid=%s err=%v", userID, sid, err)
		return errcode.ErrInternal
	}
	return nil
}
