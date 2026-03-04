package service

import (
	"context"
	"errors"
	"lesson10/internal/model"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/repository"
	"log"
	"time"

	"gorm.io/gorm"
)

type FavoriteService struct {
	favoriteRepo *repository.FavoriteRepo
	postRepo     *repository.PostRepo
}

func NewFavoriteService(favoriteRepo *repository.FavoriteRepo, postRepo *repository.PostRepo) *FavoriteService {
	return &FavoriteService{
		favoriteRepo: favoriteRepo,
		postRepo:     postRepo,
	}
}

// ToggleFavoriteService 切换收藏状态
func (r *FavoriteService) ToggleFavoriteService(ctx context.Context, uid uint, targetType uint8, targetID uint) (*bool, error) {
	const maxRetries = 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		var fav model.Favorite
		err := r.favoriteRepo.FindFav(ctx, uid, targetType, targetID, &fav)

		if err == nil {
			// 已收藏 → 取消
			result := r.favoriteRepo.DeleteFav(ctx, uid, targetType, targetID)
			if result.Error != nil {
				log.Printf("delete favorite failed (attempt %d): %v", attempt, result.Error)
				time.Sleep(50 * time.Millisecond)
				continue
			}
			log.Printf("delete favorite rows affected: %d", result.RowsAffected)
			return &[]bool{false}[0], nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("first favorite failed (attempt %d): %v", attempt, err)
			time.Sleep(50 * time.Millisecond)
			continue
		}

		// 未收藏 → 添加
		newFav := model.Favorite{
			UserID:     uid,
			TargetType: targetType,
			TargetID:   targetID,
		}
		if err = r.favoriteRepo.CreateFav(ctx, newFav); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				log.Printf("duplicate key on create favorite (attempt %d), retrying...", attempt)
				time.Sleep(50 * time.Millisecond)
				continue
			}
			log.Printf("create favorite failed (attempt %d): %v", attempt, err)
			return nil, errcode.ErrInternal
		}

		return &[]bool{true}[0], nil
	}

	return nil, errcode.ErrInternal // 重试失败
}
