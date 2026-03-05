package repository

import (
	"context"
	"lesson10/internal/dto"
	"lesson10/internal/model"

	"gorm.io/gorm"
)

type CommentRepository interface {
	FindCommentByID(ctx context.Context, commentID uint, comment *model.Comment) error
	ExistsByID(ctx context.Context, id uint) (bool, error)
	GetAndScanAuthorID(ctx context.Context, id uint) (uint, error)
	FindParentID(ctx context.Context, parent *model.Comment, req *dto.PostCommentRequest) error
	CreateComment(ctx context.Context, comment model.Comment) error
	GetAuthorID(ctx context.Context, req *dto.PostCommentRequest, AuthorID *uint)
	GetAuthorIDByComment(ctx context.Context, targetID uint, comment *model.Comment) error
	CountRootComments(ctx context.Context, targetType uint8, targetID uint) (int64, error)
	ListRootComments(ctx context.Context, req *dto.GetCommentsReq) ([]model.Comment, error)
	FindTargetComment(ctx context.Context, subs *[]model.Comment, parent uint) error
	DeleteComment(ctx context.Context, comment model.Comment) error
	DeleteSubComments(ctx context.Context, parentID uint)
}
type commentRepo struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) CommentRepository {
	return &commentRepo{db: db}
}

func (r *commentRepo) FindCommentByID(ctx context.Context, commentID uint, comment *model.Comment) error {
	err := r.db.WithContext(ctx).Where("id = ? AND is_deleted = 0", commentID).First(comment).Error
	return err
}

func (r *commentRepo) ExistsByID(ctx context.Context, id uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("id = ? AND is_deleted = 0", id).
		Count(&count).Error
	return count > 0, err
}

func (r *commentRepo) GetAndScanAuthorID(ctx context.Context, id uint) (uint, error) {
	var authorID uint
	err := r.db.WithContext(ctx).
		Model(&model.Comment{}).
		Select("author_id").
		Where("id = ? AND is_deleted = 0", id).
		Scan(&authorID).Error
	return authorID, err
}

func (r *commentRepo) FindParentID(ctx context.Context, parent *model.Comment, req *dto.PostCommentRequest) error {
	err := r.db.WithContext(ctx).Select("depth,target_id").
		Where("id = ? AND is_deleted = 0", req.TargetID).
		First(parent).Error
	return err
}

func (r *commentRepo) CreateComment(ctx context.Context, comment model.Comment) error {
	err := r.db.WithContext(ctx).Create(&comment).Error
	return err
}

func (r *commentRepo) GetAuthorID(ctx context.Context, req *dto.PostCommentRequest, AuthorID *uint) {
	r.db.WithContext(ctx).Model(&model.Comment{}).
		Select("author_id").
		Where("id = ? AND is_deleted = 0", req.TargetID).
		Scan(AuthorID)

}

func (r *commentRepo) GetAuthorIDByComment(ctx context.Context, targetID uint, comment *model.Comment) error {
	err := r.db.WithContext(ctx).Select("author_id").Where("id = ?", targetID).First(comment).Error
	return err
}

func (r *commentRepo) CountRootComments(ctx context.Context, targetType uint8, targetID uint) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).
		Model(&model.Comment{}).
		Where("target_type = ? AND target_id = ? AND depth = 1 AND is_deleted = 0", targetType, targetID).
		Count(&total).Error
	return total, err
}

func (r *commentRepo) ListRootComments(ctx context.Context, req *dto.GetCommentsReq) ([]model.Comment, error) {
	var comments []model.Comment

	offset := (req.Page - 1) * req.Size
	if offset < 0 {
		offset = 0
	}
	size := req.Size
	if size <= 0 || size > 50 {
		size = 20
	}

	err := r.db.WithContext(ctx).
		Where("target_type = ? AND target_id = ? AND depth = 1 AND is_deleted = 0", req.TargetType, req.TargetID).
		Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&comments).Error

	return comments, err
}

func (r *commentRepo) FindTargetComment(ctx context.Context, subs *[]model.Comment, parent uint) error {
	err := r.db.WithContext(ctx).Where("target_type = 3 AND target_id = ? AND is_deleted = 0", parent).
		Order("created_at DESC").
		Find(subs).Error
	return err
}

func (r *commentRepo) DeleteComment(ctx context.Context, comment model.Comment) error {
	err := r.db.WithContext(ctx).Model(&comment).Update("is_deleted", 1).Error
	return err
}

func (r *commentRepo) DeleteSubComments(ctx context.Context, parentID uint) {
	var subIDs []uint
	r.db.WithContext(ctx).Model(&model.Comment{}).
		Where("target_type = 3 AND target_id = ? AND is_deleted = 0", parentID).
		Pluck("id", &subIDs)

	if len(subIDs) == 0 {
		return
	}

	// 删当前层
	r.db.WithContext(ctx).Model(&model.Comment{}).
		Where("id IN ?", subIDs).
		Update("is_deleted", 1)

	// 递归下一层
	for _, id := range subIDs {
		r.DeleteSubComments(ctx, id)
	}
}
