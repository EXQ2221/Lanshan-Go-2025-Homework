package core

import (
	"errors"
	"lesson10/dao"
	"log"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func RegisterService(req RegisterRequest) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, ErrInternal
	}

	user := User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	if err := dao.DB.Create(&user).Error; err != nil {
		return nil, ErrInternal
	}
	return &user, nil
}

func LoginService(req LoginRequest) (string, *User, error) {
	var user User
	if err := dao.DB.Where("username = ? ", req.Username).First(&user).Error; err != nil {
		log.Println("login error db:", err)
		return "", nil, errors.New("username incorrect")

	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		log.Println("password incorrect:", err)
		return "", nil, errors.New("password incorrect")

	}

	token, err := GenerateToken(user.Username, user.ID, user.TokenVersion, user.Role)
	if err != nil {
		log.Println("login error tk:", err)
		return "", &user, ErrInternal

	}

	return token, &user, nil
}

func ChangePassService(req ChangePassRequest, id uint) error {
	var user User
	if err := dao.DB.Where("id = ? ", id).First(&user).Error; err != nil {
		return ErrInternal
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPass)); err != nil {
		return ErrForbidden
	}

	if req.NewPass == "" || req.OldPass == req.NewPass {
		return ErrForbidden
	}
	newHashBytes, err := bcrypt.GenerateFromPassword([]byte(req.NewPass), bcrypt.DefaultCost)
	if err != nil {
		return ErrInternal
	}
	newHash := string(newHashBytes)

	if err := dao.DB.Model(&User{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"password_hash": newHash,
			"token_version": gorm.Expr("token_version + 1"),
		}).Error; err != nil {
		return ErrInternal
	}

	return nil
}

func UpdateProfileService(req UpdateProfileRequest, id uint) error {
	updates := map[string]any{}

	if req.Profile != nil {
		updates["profile"] = strings.TrimSpace(*req.Profile)
	}

	if len(updates) == 0 {
		return nil
	}

	res := dao.DB.Model(&User{}).
		Where("id = ?", id).
		Updates(updates)

	if res.Error != nil {
		return ErrInternal
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func UpdateAvatarService(userID uint, avatarURL string) error {
	if err := dao.DB.Model(&User{}).
		Where("id = ?", userID).
		Update("avatar_url", avatarURL).Error; err != nil {
		return ErrInternal
	}
	return nil
}

func CreatePostService(req *CreatePostRequest, authorID uint) (*Post, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrBadRequest
	}

	if strings.TrimSpace(req.Content) == "" {
		return nil, ErrBadRequest
	}

	p := &Post{
		Type:     PostType(req.Type),
		AuthorID: authorID,
		Title:    title,
		Content:  req.Content,
		Status:   req.Status,
	}

	if err := dao.DB.Create(p).Error; err != nil {
		return nil, ErrInternal
	}
	return p, nil
}

func ListPostsService(q ListPostsQuery) ([]PostListItem, int64, error) {
	page := q.Page
	if page == 0 {
		page = 1
	}
	ps := q.PageSize
	if ps == 0 {
		ps = 20
	}
	offset := (page - 1) * ps

	db := dao.DB.Model(&Post{}).
		Where("is_deleted = 0 AND status = 0")

	if q.Type > 0 {
		db = db.Where("type = ?", q.Type)
	}

	// 关键词搜索
	if q.Keyword != "" {
		keyword := "%" + q.Keyword + "%" // 前后加 % 实现模糊搜索
		db = db.Where("title LIKE ? OR content LIKE ?", keyword, keyword)
	}

	// 总数
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, ErrInternal
	}

	// 查询列表
	var rows []PostListItem
	if err := db.Select("id, type, author_id, title, created_at, updated_at").
		Order("created_at DESC").
		Limit(ps).
		Offset(offset).
		Scan(&rows).Error; err != nil {
		return nil, 0, ErrInternal
	}

	return rows, total, nil
}

func GetPostService(id uint) (*PostDetailResp, error) {
	var p Post
	if err := dao.DB.Where("id = ? AND is_deleted = 0", id).Preload("Author").First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	if p.Status == 1 {
		if p.AuthorID != id {
			return nil, ErrUnauthorized
		}
	}
	return &PostDetailResp{
		ID:         p.ID,
		Type:       uint8(p.Type),
		AuthorID:   p.AuthorID,
		AuthorName: p.Author.Username,
		Title:      p.Title,
		Content:    p.Content,
		LikeCount:  p.LikeCount,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}, nil
}
func PostCommentService(id uint, req *PostCommentRequest) (*Comment, error) {
	var pDepth uint8 = 0
	if req.TargetType == 3 {
		var parent Comment
		err := dao.DB.Select("depth,target_id").
			Where("id = ? AND is_deleted = 0", req.TargetID).
			First(&parent).Error

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("fail to find comment")
		}
		if err != nil {
			return nil, ErrInternal
		}

		pDepth = parent.Depth

		if pDepth >= 7 {
			return nil, err
		}
	} else {
		//防止乱传数据
		var count int64
		if err := dao.DB.Model(&Post{}).
			Where("id = ? AND is_deleted = 0", req.TargetID).
			Count(&count).Error; err != nil {
			return nil, ErrInternal
		}

		if count == 0 {
			return nil, errors.New("fail to find post or question")
		}
	}

	comment := Comment{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		AuthorID:   id,
		Content:    req.Content,
		Depth:      pDepth + 1,
	}

	if err := dao.DB.Create(&comment).Error; err != nil {
		return nil, ErrInternal
	}

	//通知
	var receiverID uint
	var notifyType uint8 = 1
	var content string
	if req.TargetType == 3 {
		// 二级及以上：通知直接父评论的作者
		var parentAuthorID uint
		dao.DB.Model(&Comment{}).
			Select("author_id").
			Where("id = ?", req.TargetID).
			Scan(&parentAuthorID)

		if parentAuthorID != 0 && parentAuthorID != id {
			receiverID = parentAuthorID
			content = "有人回复了你的评论"
		}
	} else {
		// 一级评论：通知帖子/问题作者
		var authorID uint
		err := dao.DB.Model(&Post{}).
			Select("author_id").
			Where("id = ?", req.TargetID).
			Scan(&authorID)

		if err == nil && authorID != 0 && authorID != id {
			receiverID = authorID
			if req.TargetType == 1 {
				content = "有人评论了你的文章"
			} else {
				content = "有人回答了你的问题"
			}
		}
	}

	if receiverID != 0 {
		notification := Notification{
			UserID:     receiverID,
			Type:       notifyType,
			ActorID:    &id,
			TargetType: (*uint8)(&req.TargetType),
			TargetID:   &req.TargetID,
			Content:    content,
		}
		_ = dao.DB.Create(&notification).Error
	}

	return &comment, nil
}

func GetCommentsService(req *GetCommentsReq) (*GetCommentsResp, error) {
	var comments []Comment

	offset := (req.Page - 1) * req.Size
	if offset < 0 {
		offset = 0
	}
	if req.Size <= 0 || req.Size > 50 {
		req.Size = 20
	}

	// 只查一级评论
	query := dao.DB.Model(&Comment{}).Where("target_type = ? AND target_id = ? AND depth = 1 AND is_deleted = 0",
		req.TargetType, req.TargetID)

	var total int64
	query.Count(&total)

	// 分页 排序
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.Size).
		Find(&comments).Error

	if err != nil {
		log.Printf("查询评论失败: %v", err)
		return nil, ErrInternal
	}

	items := make([]CommentItem, len(comments))

	for i, c := range comments {
		var author User
		dao.DB.Select("username").
			Where("id = ?", c.AuthorID).
			First(&author)
		items[i] = CommentItem{
			ID:         c.ID,
			AuthorID:   c.AuthorID,
			AuthorName: author.Username,
			Content:    c.Content,
			Depth:      c.Depth,
			CreatedAt:  c.CreatedAt,
		}
	}

	return &GetCommentsResp{
		Comments: items,
		Total:    total,
		Page:     req.Page,
		Size:     req.Size,
	}, nil
}

// 二级以上评论
func GetReplies(parentID uint, page, size int) ([]Comment, int64, error) {
	offset := (page - 1) * size

	var total int64
	dao.DB.Model(&Comment{}).
		Where("target_type = 3 AND target_id = ? AND is_deleted = 0", parentID).
		Count(&total)

	var replies []Comment
	err := dao.DB.Where("target_type = 3 AND target_id = ? AND is_deleted = 0", parentID).
		Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&replies).Error

	if err != nil {
		return nil, 0, err
	}

	return replies, total, nil
}

func UpdatePostService(PostID uint64, id uint, req *UpdatePostRequest) error {
	var post Post
	err := dao.DB.Where("id = ? AND is_deleted = 0", PostID).First(&post).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("fail to find post")
	}
	if err != nil {
		return ErrInternal
	}

	if post.AuthorID != id {
		return ErrUnauthorized
	}

	updates := map[string]interface{}{}
	if req.Title == "" {
		updates["title"] = req.Title
	}

	if req.Content == "" {
		updates["content"] = req.Content
	}

	if len(updates) == 0 {
		return errors.New("please update at least one")
	}

	updates["updateTime"] = time.Now()

	return dao.DB.Model(&post).Updates(updates).Error
}

func DeletePostService(postID, uid uint, role uint) error {
	var post Post
	err := dao.DB.Where("id = ? AND is_deleted = 0", postID).First(&post).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("fail to find post")
	}
	if err != nil {
		return err
	}

	if post.AuthorID != uid && role != uint(RoleAdmin) {
		return ErrUnauthorized
	}

	return dao.DB.Model(&post).Update("is_deleted", 1).Error

}

func DeleteComment(commentID, uid uint, role uint) error {
	var comment Comment
	err := dao.DB.Where("id = ? AND is_deleted = 0", commentID).First(&comment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("fail to find comment")
	}
	if err != nil {
		return ErrInternal
	}

	if comment.AuthorID != uid && role != uint(RoleAdmin) {
		return ErrUnauthorized
	}

	// 如果评论是一级评论，递归删除所有子评论
	if comment.TargetType != CommentOnComment {
		deleteSubComments(comment.ID)

	}

	// 如果不是一级评论，删除自己。如果是一级评论，在递归删掉子评论后删除自己
	return dao.DB.Model(&comment).Update("is_deleted", 1).Error
}

func GetUserInfoService(currentID, id uint, page int) (*UserPublicInfo, error) {
	var user User
	err := dao.DB.Where("id = ?", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("fail to find user")
	}
	if err != nil {
		return nil, ErrInternal
	}

	var posts []Post
	var total int64
	var size = 5
	offset := (page - 1) * size

	baseQuery := dao.DB.Model(&Post{}).
		Where("author_id = ? AND is_deleted = 0 AND status = 0", id)

	baseQuery.Count(&total)

	baseQuery.Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&posts)

	// 如果当前用户是本人，额外查草稿
	var draftPosts []Post
	var draftTotal int64
	if currentID == id {
		draftQuery := dao.DB.Model(&Post{}).
			Where("author_id = ? AND is_deleted = 0 AND status = 1", id)

		draftQuery.Count(&draftTotal)

		draftQuery.Order("created_at DESC").
			Offset(offset).
			Limit(size).
			Find(&draftPosts)

		// 可以合并到 posts，或者分开返回（推荐分开）
		// 这里示例合并到同一个数组（已发布在前，草稿在后）
		posts = append(posts, draftPosts...)
		total += draftTotal
	}

	postSummaries := make([]PostSummary, len(posts))
	for i, p := range posts {
		postSummaries[i] = PostSummary{
			ID:        p.ID,
			Title:     p.Title,
			CreatedAt: p.CreatedAt,
			Status:    p.Status,
		}
	}

	isVIP := user.VIPExpiresAt != nil && time.Now().Before(*user.VIPExpiresAt)

	userPublicInfo := UserPublicInfo{
		ID:           user.ID,
		Username:     user.Username,
		AvatarURL:    user.AvatarURL,
		Profile:      user.Profile,
		Role:         user.Role,
		IsVIP:        isVIP,
		VIPExpiresAt: user.VIPExpiresAt,
		Posts:        postSummaries,
		PostTotal:    total,
		Page:         page,
		Size:         size,
	}

	return &userPublicInfo, nil
}

func FollowUserService(followerID, followeeID uint) error {
	var count int64
	dao.DB.Model(&User{}).
		Where("id = ?", followeeID).
		Count(&count)
	if count == 0 {
		return ErrNotFound
	}

	var existing UserFollow
	err := dao.DB.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		First(&existing).Error
	if err == nil {
		return errors.New("has followed")
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	} else if err != nil {
		return ErrInternal
	}

	follow := UserFollow{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}

	return dao.DB.Create(&follow).Error
}

func UnfollowUserService(followerID, followeeID uint) error {
	result := dao.DB.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		Delete(&UserFollow{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("has not followed")
	}
	return nil
}

func ToggleReactionService(uid uint, targetType uint8, targetID uint) (*bool, error) {
	var isLiked *bool // 先定义一个指针变量，用于存储最终状态

	err := dao.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 校验目标存在（用 tx）
		var count int64
		switch targetType {
		case 1, 2:
			tx.Model(&Post{}).
				Where("id = ? AND is_deleted = 0", targetID).
				Count(&count)
		case 3:
			tx.Model(&Comment{}).
				Where("id = ? AND is_deleted = 0", targetID).
				Count(&count)
		default:
			return ErrBadRequest
		}
		if count == 0 {
			return ErrNotFound
		}

		// 2. 检查是否已点赞
		var reaction Reaction
		err := tx.Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).
			First(&reaction).Error

		if err == nil {
			// 已点赞 → 取消
			if err := tx.Delete(&reaction).Error; err != nil {
				return ErrInternal
			}
			if err := decrementLikeCountTx(tx, targetType, targetID); err != nil {
				return ErrInternal
			}
			temp := false
			isLiked = &temp
			return nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInternal
		}

		// 未点赞 → 添加
		newReaction := Reaction{
			UserID:     uid,
			TargetType: ReactionTargetType(targetType),
			TargetID:   targetID,
		}
		if err := tx.Create(&newReaction).Error; err != nil {
			return ErrInternal
		}
		if err := incrementLikeCountTx(tx, targetType, targetID); err != nil {
			return ErrInternal
		}
		temp := true
		isLiked = &temp
		return nil
	})

	if err != nil {
		return nil, err // 事务失败，返回错误
	}

	// 事务成功，返回操作后的 isLiked
	return isLiked, nil
}

// 可选：发通知给目标作者
// go sendLikeNotification(uid, targetType, targetID)

// ToggleFavoriteService 切换收藏状态，返回操作后的“是否已收藏”
func ToggleFavoriteService(uid uint, targetType uint8, targetID uint) (*bool, error) {
	var isFavorited *bool

	err := dao.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 校验目标是否存在
		var count int64
		switch targetType {
		case 1, 2: // 文章或问题（都在 posts 表）
			tx.Model(&Post{}).
				Where("id = ? AND is_deleted = 0", targetID).
				Count(&count)
		default:
			return ErrBadRequest
		}
		if count == 0 {
			return ErrNotFound
		}

		// 2. 检查是否已收藏
		var fav Favorite
		err := tx.Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).
			First(&fav).Error

		if err == nil {
			// 已收藏 → 取消收藏
			if err := tx.Delete(&fav).Error; err != nil {
				return ErrInternal
			}
			isFavorited = new(bool) // false
			*isFavorited = false
			return nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInternal
		}

		// 未收藏 → 添加收藏
		newFav := Favorite{
			UserID:     uid,
			TargetType: FavoriteTargetType(targetType),
			TargetID:   targetID,
		}
		if err := tx.Create(&newFav).Error; err != nil {
			return ErrInternal
		}

		isFavorited = new(bool)
		*isFavorited = true
		return nil
	})

	return isFavorited, err
}

// GetFollowListService 获取关注/粉丝列表
func GetFollowListService(targetUserID uint, listType string, currentUserID uint, page, size int) ([]FollowUserInfo, int64, error) {
	offset := (page - 1) * size // 计算偏移量

	var total int64
	var followIDs []uint

	query := dao.DB.Model(&UserFollow{})

	if listType == "followers" {
		query = query.Where("followee_id = ?", targetUserID)
	} else if listType == "following" {
		query = query.Where("follower_id = ?", targetUserID)
	} else {
		return nil, 0, errors.New("invalid list type")
	}

	// 总数（不受分页影响）
	query.Count(&total)

	// 获取对方用户ID列表 + 分页（关键：加 Offset 和 Limit）
	if listType == "followers" {
		query.Select("follower_id").
			Offset(offset).
			Limit(size).
			Pluck("follower_id", &followIDs)
	} else {
		query.Select("followee_id").
			Offset(offset).
			Limit(size).
			Pluck("followee_id", &followIDs)
	}

	if len(followIDs) == 0 {
		return []FollowUserInfo{}, total, nil
	}

	var userList []struct {
		ID        uint   `gorm:"column:id"`
		Username  string `gorm:"column:username"`
		AvatarURL string `gorm:"column:avatar_url"`
		Profile   string `gorm:"column:profile"`
	}
	dao.DB.Table("users").
		Select("id, username, avatar_url, profile").
		Where("id IN ?", followIDs).
		Find(&userList)

	userMap := make(map[uint]struct {
		ID        uint
		Username  string
		AvatarURL string
		Profile   string
	})
	for _, u := range userList {
		userMap[u.ID] = struct {
			ID        uint
			Username  string
			AvatarURL string
			Profile   string
		}(u)
	}

	result := make([]FollowUserInfo, len(followIDs))
	for i, id := range followIDs {
		u := userMap[id]
		result[i] = FollowUserInfo{
			ID:         u.ID,
			Username:   u.Username,
			AvatarURL:  u.AvatarURL,
			Profile:    u.Profile,
			IsFollowed: false,
		}

		// 仅登录状态判断
		if currentUserID > 0 {
			var count int64
			dao.DB.Model(&UserFollow{}).
				Where("follower_id = ? AND followee_id = ?", currentUserID, id).
				Count(&count)
			result[i].IsFollowed = count > 0 // count == 1为true 0 就是false
		}
	}

	return result, total, nil
}

func GetNotifications(uid uint, page, size int, unreadOnly bool) ([]NotificationItem, int64, error) {
	offset := (page - 1) * size

	var total int64
	dao.DB.Model(&Notification{}).
		Where("user_id = ?", uid).
		Count(&total)

	query := dao.DB.Model(&Notification{}).
		Where("user_id = ?", uid).
		Order("created_at DESC").
		Offset(offset).
		Limit(size)

	if unreadOnly {
		query = query.Where("is_read = 0")
	}

	var notifications []Notification
	if err := query.Find(&notifications).Error; err != nil {
		return nil, 0, ErrInternal
	}

	// 批量查触发者用户名（可选）
	actorIDs := make([]uint, 0, len(notifications))
	for _, n := range notifications {
		if n.ActorID != nil {
			actorIDs = append(actorIDs, *n.ActorID)
		}
	}

	var actors []User
	if len(actorIDs) > 0 {
		dao.DB.Select("id, username").
			Where("id IN ?", actorIDs).
			Find(&actors)
	}

	actorMap := make(map[uint]string)
	for _, a := range actors {
		actorMap[a.ID] = a.Username
	}

	items := make([]NotificationItem, len(notifications))
	for i, n := range notifications {
		var actorID uint = 0
		if n.ActorID != nil {
			actorID = *n.ActorID
		} else {
			actorID = 0
		}

		if n.ActorID != nil {
			actorID = *n.ActorID
		}

		var targetType uint8 = 0
		if n.TargetType != nil {
			targetType = *n.TargetType
		}

		var targetID uint = 0
		if n.TargetID != nil {
			targetID = *n.TargetID
		}
		items[i] = NotificationItem{
			ID:         n.ID,
			Type:       n.Type,
			ActorID:    actorID,
			ActorName:  actorMap[actorID],
			TargetType: targetType,
			TargetID:   targetID,
			Content:    n.Content,
			IsRead:     n.IsRead == 1,
			CreatedAt:  n.CreatedAt,
		}
	}

	return items, total, nil
}
