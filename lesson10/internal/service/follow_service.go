package service

import (
	"context"
	"lesson10/internal/dto"
	"lesson10/internal/model"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/repository"
	"log"
	"strings"
)

type FollowService struct {
	userRepo   repository.UserRepository
	followRepo repository.FollowRepository
}

func NewFollowService(followRepo repository.FollowRepository, userRepo repository.UserRepository) *FollowService {
	return &FollowService{
		followRepo: followRepo,
		userRepo:   userRepo,
	}
}

func (r *FollowService) FollowUserService(ctx context.Context, followerID, followeeID uint) error {
	// 1. 校验被关注者存在
	exists, err := r.userRepo.ExistsByUserID(ctx, followeeID)
	if err != nil {
		return errcode.ErrInternal
	}
	if !exists {
		return errcode.ErrNotFound
	}

	// 2. 先尝试插入（关注）
	follow := model.UserFollow{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}

	createErr := r.followRepo.CreateFollow(ctx, follow)
	if createErr == nil {
		return nil
	}

	// 插入失败 → 检查是否重复键
	if strings.Contains(createErr.Error(), "Duplicate entry") {
		return errcode.ErrHasFollowed
	}

	// 其他错误
	log.Printf("create follow failed: %v", createErr)
	return errcode.ErrInternal
}

func (r *FollowService) UnfollowUserService(ctx context.Context, followerID, followeeID uint) error {
	result := r.followRepo.DeleteFollow(ctx, followerID, followeeID)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errcode.ErrHasNotFollowed
	}

	return nil
}

// GetFollowListService 获取关注/粉丝列表
func (r *FollowService) GetFollowListService(ctx context.Context, targetUserID uint, listType string, currentUserID uint, page, size int) ([]dto.FollowUserInfo, int64, error) {
	if listType != "followers" && listType != "following" {
		return nil, 0, errcode.ErrInvalidListType
	}

	offset := (page - 1) * size
	if offset < 0 {
		offset = 0
	}
	if size <= 0 || size > 50 {
		size = 20
	}

	// 1. 总数
	total, err := r.followRepo.CountFollows(ctx, targetUserID, listType == "followers")
	if err != nil {
		return nil, 0, errcode.ErrInternal
	}

	// 2. 获取对方用户 ID 列表
	followIDs, err := r.followRepo.ListFollowIDs(ctx, targetUserID, listType == "followers", offset, size)
	if err != nil {
		return nil, 0, errcode.ErrInternal
	}

	if len(followIDs) == 0 {
		return []dto.FollowUserInfo{}, total, nil
	}

	// 3. 批量查用户信息（用户名、头像、简介）
	userMap, err := r.userRepo.BatchGetUserBasicInfo(ctx, followIDs)
	if err != nil {
		log.Printf("批量查用户失败: %v", err)
		// 降级：继续返回空用户名
	}

	// 4. 批量查当前用户是否关注这些人
	isFollowedMap := make(map[uint]bool)
	if currentUserID > 0 {
		isFollowedMap, err = r.followRepo.BatchIsFollowing(ctx, currentUserID, followIDs)
		if err != nil {
			log.Printf("批量查关注状态失败: %v", err)
			// 降级：全部设为 false
		}
	}

	// 5. 组装 DTO
	result := make([]dto.FollowUserInfo, len(followIDs))
	for i, uid := range followIDs {
		u, ok := userMap[uid]
		username := ""
		avatar := ""
		profile := ""
		if ok {
			username = u.Username
			avatar = u.AvatarURL
			profile = u.Profile
		}

		result[i] = dto.FollowUserInfo{
			ID:         uid,
			Username:   username,
			AvatarURL:  avatar,
			Profile:    profile,
			IsFollowed: isFollowedMap[uid],
		}
	}

	return result, total, nil
}
