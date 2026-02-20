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

func incrementLikeCountTx(tx *gorm.DB, targetType uint8, targetID uint) error {
	switch targetType {
	case 1, 2:
		return tx.Model(&Post{}).
			Where("id = ?", targetID).
			Update("like_count", gorm.Expr("like_count + ?", 1)).Error
	case 3:
		return tx.Model(&Comment{}).
			Where("id = ?", targetID).
			Update("like_count", gorm.Expr("like_count + ?", 1)).Error
	default:
		return nil
	}
}

func decrementLikeCountTx(tx *gorm.DB, targetType uint8, targetID uint) error {
	switch targetType {
	case 1, 2:
		return tx.Model(&Post{}).
			Where("id = ? AND like_count > 0", targetID).
			Update("like_count", gorm.Expr("like_count - ?", 1)).Error
	case 3:
		return tx.Model(&Comment{}).
			Where("id = ? AND like_count > 0", targetID).
			Update("like_count", gorm.Expr("like_count - ?", 1)).Error
	default:
		return nil
	}
}
