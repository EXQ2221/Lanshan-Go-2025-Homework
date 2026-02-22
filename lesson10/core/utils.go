package core

import (
	"lesson10/dao"

	"gorm.io/gorm"
)

func deleteSubComments(parentID uint) {
	var subIDs []uint
	dao.DB.Model(&Comment{}).
		Where("target_type = 3 AND target_id = ? AND is_deleted = 0", parentID).
		Pluck("id", &subIDs)

	if len(subIDs) == 0 {
		return
	}

	// 删当前层
	dao.DB.Model(&Comment{}).
		Where("id IN ?", subIDs).
		Update("is_deleted", 1)

	// 递归下一层
	for _, id := range subIDs {
		deleteSubComments(id)
	}
}

func incrementLikeCount(targetType uint8, targetID uint) {
	switch targetType {
	case 1, 2:
		dao.DB.Model(&Post{}).
			Where("id = ?", targetID).
			Update("like_count", gorm.Expr("like_count + 1"))
	case 3:
		dao.DB.Model(&Comment{}).
			Where("id = ?", targetID).
			Update("like_count", gorm.Expr("like_count + 1"))
	}
}

func decrementLikeCount(targetType uint8, targetID uint) {
	switch targetType {
	case 1, 2:
		dao.DB.Model(&Post{}).
			Where("id = ?", targetID).
			Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
	case 3:
		dao.DB.Model(&Comment{}).
			Where("id = ?", targetID).
			Update("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
	}
}
