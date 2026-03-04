package service

import (
	"context"
	"errors"
	"lesson10/internal/dto"
	"lesson10/internal/model"
	"lesson10/internal/pkg/errcode"
	"lesson10/internal/repository"
	"log"

	"gorm.io/gorm"
)

type CommentService struct {
	userRepo         *repository.UserRepo
	postRepo         *repository.PostRepo
	commentRepo      *repository.CommentRepo
	notificationRepo *repository.NotificationRepo
	reactionRepo     *repository.ReactionRepo
}

func NewCommentService(userRepo *repository.UserRepo, postRepo *repository.PostRepo, commentRepo *repository.CommentRepo, notificationRepo *repository.NotificationRepo, reactionRepo *repository.ReactionRepo) *CommentService {
	return &CommentService{
		userRepo:         userRepo,
		postRepo:         postRepo,
		commentRepo:      commentRepo,
		notificationRepo: notificationRepo,
		reactionRepo:     reactionRepo,
	}
}

func (r *CommentService) PostCommentService(ctx context.Context, id uint, req *dto.PostCommentRequest) (*model.Comment, error) {
	var pDepth uint8 = 0
	if req.TargetType == 3 {
		var parent model.Comment
		err := r.commentRepo.FindParentID(ctx, &parent, req)

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		if err != nil {
			return nil, errcode.ErrInternal
		}

		pDepth = parent.Depth

		if pDepth >= 7 {
			return nil, err
		}

	} else {
		exists, err := r.postRepo.ExistsPostByID(ctx, req.TargetID)
		if err != nil {
			return nil, errcode.ErrInternal
		}
		if !exists {
			return nil, errcode.ErrNotFound
		}
	}

	comment := model.Comment{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		AuthorID:   id,
		Content:    req.Content,
		Depth:      pDepth + 1,
	}

	if err := r.commentRepo.CreateComment(ctx, comment); err != nil {
		return nil, errcode.ErrInternal
	}

	//通知
	var receiverID uint
	var notifyType uint8 = 1
	var content string
	if req.TargetType == 3 {
		// 二级及以上：通知直接父评论的作者
		var parentAuthorID uint
		r.commentRepo.GetAuthorID(ctx, req, &parentAuthorID)

		if parentAuthorID != 0 && parentAuthorID != id {
			receiverID = parentAuthorID
			content = "有人回复了你的评论"
		}
	} else {
		// 一级评论：通知帖子/问题作者
		var authorID uint
		r.commentRepo.GetAuthorID(ctx, req, &authorID)

		if authorID != 0 && authorID != id {
			receiverID = authorID
			if req.TargetType == 1 {
				content = "有人评论了你的文章"
			} else {
				content = "有人回答了你的问题"
			}
		}
	}

	if receiverID != 0 {
		notification := &model.Notification{
			UserID:     receiverID,
			Type:       notifyType,
			ActorID:    &id,
			TargetType: (*uint8)(&req.TargetType),
			TargetID:   &req.TargetID,
			Content:    content,
		}
		_ = r.notificationRepo.CreateNotification(ctx, notification)
	}

	return &comment, nil
}

func (r *CommentService) GetCommentsService(ctx context.Context, req *dto.GetCommentsReq) (*dto.GetCommentsResp, error) {
	var comments []model.Comment

	offset := (req.Page - 1) * req.Size
	if offset < 0 {
		offset = 0
	}
	if req.Size <= 0 || req.Size > 50 {
		req.Size = 20
	}

	// 只查一级评论
	total, err := r.commentRepo.CountRootComments(ctx, req.TargetType, req.TargetID)
	if err != nil {
		log.Printf("查询评论总数失败: %v", err)
		return nil, errcode.ErrInternal
	}

	// 分页 排序
	comments, err = r.commentRepo.ListRootComments(ctx, req)
	if err != nil {
		log.Printf("查询评论列表失败: %v", err)
		return nil, errcode.ErrInternal
	}

	if len(comments) == 0 {
		return &dto.GetCommentsResp{
			Comments: []dto.CommentItem{},
			Total:    total,
			Page:     req.Page,
			Size:     req.Size,
		}, nil
	}

	// 批量查作者用户名（解决 N+1）
	authorIDs := make([]uint, 0, len(comments))
	for _, c := range comments {
		authorIDs = append(authorIDs, c.AuthorID)
	}

	authorMap, err := r.userRepo.BatchGetAuthorUsernames(ctx, authorIDs)
	if err != nil {
		log.Printf("批量查询作者失败: %v", err)
	}

	// 组装 DTO
	items := make([]dto.CommentItem, len(comments))
	for i, c := range comments {
		username := ""
		if name, ok := authorMap[c.AuthorID]; ok {
			username = name
		}

		items[i] = dto.CommentItem{
			ID:         c.ID,
			AuthorID:   c.AuthorID,
			AuthorName: username,
			Content:    c.Content,
			Depth:      c.Depth,
			CreatedAt:  c.CreatedAt,
		}
	}

	return &dto.GetCommentsResp{
		Comments: items,
		Total:    total,
		Page:     req.Page,
		Size:     req.Size,
	}, nil
}

// GetAllReplies 查询二级以上评论
func (r *CommentService) GetAllReplies(ctx context.Context, parentID uint, currentUID uint) ([]dto.CommentItem, int64, error) {
	var allReplies []model.Comment
	// 递归函数
	var fetchReplies func(parent uint) error
	fetchReplies = func(parent uint) error {
		var subs []model.Comment
		err := r.commentRepo.FindTargetComment(ctx, &subs, parent)
		if err != nil {
			return err
		}

		for _, sub := range subs {
			allReplies = append(allReplies, sub)
			if err = fetchReplies(sub.ID); err != nil { //递归调用
				return err
			}
		}
		return nil
	}

	if err := fetchReplies(parentID); err != nil {
		return nil, 0, err
	}
	total := int64(len(allReplies))
	// 批量查作者名
	authorIDs := make([]uint, 0, len(allReplies))
	for _, c := range allReplies {
		authorIDs = append(authorIDs, c.AuthorID)
	}

	authorMap, err := r.userRepo.BatchGetAuthorUsernames(ctx, authorIDs)
	if err != nil {
		log.Printf("批量查询作者失败: %v", err)
	}

	// 批量查 is_liked（当前用户是否点赞这些评论）
	commentIDs := make([]uint, 0, len(allReplies))
	for _, c := range allReplies {
		commentIDs = append(commentIDs, c.ID)
	}

	likedMap := make(map[uint]bool)
	if currentUID > 0 && len(commentIDs) > 0 {
		likedMap, err = r.reactionRepo.BatchCheckLikedByUser(ctx, currentUID, commentIDs)
	}

	// 转成返回结构
	items := make([]dto.CommentItem, len(allReplies))
	for i, c := range allReplies {
		isLiked := false // ← 这里定义 isLiked
		if currentUID > 0 {
			isLiked = likedMap[c.ID]
		}
		items[i] = dto.CommentItem{
			ID:         c.ID,
			AuthorID:   c.AuthorID,
			AuthorName: authorMap[c.AuthorID],
			Content:    c.Content,
			Depth:      c.Depth,
			CreatedAt:  c.CreatedAt,
			LikeCount:  int(c.LikeCount),
			IsLiked:    isLiked,
		}
	}
	return items, total, nil
}

func (r *CommentService) DeleteComment(ctx context.Context, commentID, uid uint, role uint) error {
	var comment model.Comment
	err := r.commentRepo.FindCommentByID(ctx, commentID, &comment)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrNotFound
	}
	if err != nil {
		return errcode.ErrInternal
	}

	if comment.AuthorID != uid && role != uint(model.RoleAdmin) {
		return errcode.ErrUnauthorized
	}

	// 如果评论是一级评论，递归删除所有子评论
	if comment.TargetType != model.CommentOnComment {
		r.commentRepo.DeleteSubComments(ctx, comment.ID)

	}

	// 如果不是一级评论，删除自己。如果是一级评论，在递归删掉子评论后删除自己
	return r.commentRepo.DeleteComment(ctx, comment)
}
