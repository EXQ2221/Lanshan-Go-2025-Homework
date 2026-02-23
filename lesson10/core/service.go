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

	// 基础查询：只查未删除、已发布的帖子
	db := dao.DB.Model(&Post{}).
		Where("is_deleted = 0 AND status = 0")

	if q.Type > 0 {
		db = db.Where("type = ?", q.Type)
	}

	// 关键词搜索（标题或内容）
	if q.Keyword != "" {
		keyword := "%" + q.Keyword + "%"
		db = db.Where("title LIKE ? OR content LIKE ?", keyword, keyword)
	}

	// 统计总数（不变）
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, ErrInternal
	}

	// 改用 JOIN 查询，同时取出作者用户名
	// 注意：这里使用 Table + Joins + Select，性能更好，也能避免 Preload 的 N+1 问题
	var rows []struct {
		ID         uint      `gorm:"column:id"`
		Type       uint8     `gorm:"column:type"`
		AuthorID   uint      `gorm:"column:author_id"`
		Title      string    `gorm:"column:title"`
		CreatedAt  time.Time `gorm:"column:created_at"`
		UpdatedAt  time.Time `gorm:"column:updated_at"`
		AuthorName string    `gorm:"column:username"` // 从 users 表取
	}

	err := db.Table("posts p").
		Joins("LEFT JOIN users u ON p.author_id = u.id").
		Select("p.id, p.type, p.author_id, p.title, p.created_at, p.updated_at, u.username").
		Order("p.updated_at DESC"). // ← 改用 updated_at 排序（最新活跃的排前面）
		Limit(ps).
		Offset(offset).
		Scan(&rows).Error

	if err != nil {
		return nil, 0, ErrInternal
	}

	// 转换成前端需要的 PostListItem 结构
	items := make([]PostListItem, len(rows))
	for i, r := range rows {
		items[i] = PostListItem{
			ID:         r.ID,
			Type:       r.Type,
			AuthorID:   r.AuthorID,
			AuthorName: r.AuthorName, // ← 这里就有用户名了
			Title:      r.Title,
			CreateAt:   r.CreatedAt,
			UpdatedAt:  r.UpdatedAt,
		}
	}

	return items, total, nil
}
func GetPostService(currentID, id uint) (*PostDetailResp, error) {
	var p Post
	if err := dao.DB.Where("id = ? AND is_deleted = 0", id).Preload("Author").First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrInternal
	}

	if p.Status == 1 {
		if p.AuthorID != currentID {
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
		Status:     p.Status,
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
		db := dao.DB.Model(&Post{}).
			Select("author_id").
			Where("id = ?", req.TargetID).
			Scan(&authorID)

		if db.Error == nil && authorID != 0 && authorID != id {
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
func GetAllReplies(parentID uint, currentUID uint) ([]CommentItem, int64, error) {
	var allReplies []Comment
	// 递归函数
	var fetchReplies func(parent uint) error
	fetchReplies = func(parent uint) error {
		var subs []Comment
		err := dao.DB.Where("target_type = 3 AND target_id = ? AND is_deleted = 0", parent).
			Order("created_at DESC").
			Find(&subs).Error
		if err != nil {
			return err
		}
		for _, sub := range subs {
			allReplies = append(allReplies, sub)
			if err := fetchReplies(sub.ID); err != nil {
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
	authorMap := make(map[uint]string)
	if len(authorIDs) > 0 {
		var authors []User
		dao.DB.Select("id, username").
			Where("id IN ?", authorIDs).
			Find(&authors)
		for _, a := range authors {
			authorMap[a.ID] = a.Username
		}
	}
	// 批量查 is_liked（当前用户是否点赞这些评论）
	commentIDs := make([]uint, 0, len(allReplies))
	for _, c := range allReplies {
		commentIDs = append(commentIDs, c.ID)
	}
	likedMap := make(map[uint]bool)
	if currentUID > 0 && len(commentIDs) > 0 {
		var likedComments []Reaction
		dao.DB.Where("user_id = ? AND target_type = 3 AND target_id IN ?", currentUID, commentIDs).
			Find(&likedComments)
		for _, l := range likedComments {
			likedMap[l.TargetID] = true
		}
	}
	// 转成返回结构
	items := make([]CommentItem, len(allReplies))
	for i, c := range allReplies {
		isLiked := false // ← 这里定义 isLiked
		if currentUID > 0 {
			isLiked = likedMap[c.ID]
		}
		items[i] = CommentItem{
			ID:         c.ID,
			AuthorID:   c.AuthorID,
			AuthorName: authorMap[c.AuthorID],
			Content:    c.Content,
			Depth:      c.Depth,
			CreatedAt:  c.CreatedAt,
			LikeCount:  int(c.LikeCount),
			IsLiked:    isLiked, // 当前用户是否点赞这条评论
		}
	}
	return items, total, nil
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
	if req.Title != "" {
		updates["title"] = strings.TrimSpace(req.Title)
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	updates["updated_at"] = time.Now()

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

	var followingCount int64
	dao.DB.Model(&UserFollow{}).
		Where("follower_id = ?", id). // 我关注的人
		Count(&followingCount)

	var followersCount int64
	dao.DB.Model(&UserFollow{}).
		Where("followee_id = ?", id). // 关注我的人
		Count(&followersCount)

	isFollowed := false
	if currentID > 0 && currentID != id {
		var cnt int64
		dao.DB.Model(&UserFollow{}).
			Where("follower_id = ? AND followee_id = ?", currentID, id).
			Count(&cnt)
		isFollowed = cnt > 0
	}

	userPublicInfo := UserPublicInfo{
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

func FollowUserService(followerID, followeeID uint) error {
	// 1. 校验被关注者存在
	var count int64
	dao.DB.Model(&User{}).
		Where("id = ?", followeeID).
		Count(&count)
	if count == 0 {
		return ErrNotFound
	}

	// 2. 先尝试插入（关注）
	follow := UserFollow{
		FollowerID: followerID,
		FolloweeID: followeeID,
	}

	createErr := dao.DB.Create(&follow).Error
	if createErr == nil {
		return nil
	}

	// 插入失败 → 检查是否重复键
	if strings.Contains(createErr.Error(), "Duplicate entry") {
		return errors.New("has followed")
	}

	// 其他错误
	log.Printf("create follow failed: %v", createErr)
	return ErrInternal
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

// ToggleReactionService 切换点赞状态，返回操作后的“是否已点赞”
func ToggleReactionService(uid uint, targetType uint8, targetID uint) (*bool, error) {

	const maxRetries = 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(50 * time.Millisecond) // 等待前操作完成
		}

		if targetType == 1 || targetType == 2 {
			var post Post
			err := dao.DB.Where("id = ? AND is_deleted = 0", targetID).First(&post).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrNotFound
			}
			if err != nil {
				return nil, ErrInternal
			}

			// 草稿状态（status = 1）只能作者本人点赞
			if post.Status == 1 && post.AuthorID != uid {
				return nil, ErrUnauthorized
			}
		}
		// 检查是否已点赞
		var reaction Reaction
		err := dao.DB.Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).
			First(&reaction).Error

		if err == nil {
			// 已点赞 → 取消（直接用唯一条件删除）
			result := dao.DB.Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).
				Delete(&Reaction{})

			if result.Error != nil {
				log.Printf("delete failed (attempt %d): %v", attempt, result.Error)
				continue
			}

			log.Printf("delete rows affected: %d", result.RowsAffected)

			decrementLikeCount(targetType, targetID)
			isLiked := false
			return &isLiked, nil
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("first reaction failed (attempt %d): %v", attempt, err)
			continue
		}

		// 未点赞 → 添加
		newReaction := Reaction{
			UserID:     uid,
			TargetType: targetType,
			TargetID:   targetID,
		}

		if err := dao.DB.Create(&newReaction).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				log.Printf("duplicate key on create (attempt %d), retrying...", attempt)
				continue
			}
			log.Printf("create failed (attempt %d): %v", attempt, err)
			return nil, ErrInternal
		}

		if targetType == 1 || targetType == 2 || targetType == 3 {
			var receiverID uint

			if targetType == 1 || targetType == 2 {
				var post Post
				if err := dao.DB.Select("author_id").Where("id = ?", targetID).First(&post).Error; err == nil {
					receiverID = post.AuthorID
				}
			} else {
				var comment Comment
				if err := dao.DB.Select("author_id").Where("id = ?", targetID).First(&comment).Error; err == nil {
					receiverID = comment.AuthorID
				}
			}

			if receiverID != 0 && receiverID != uid {
				notifyType := uint8(2) // 2表示点赞通知
				content := "有人点赞了你的内容"
				tt := targetType
				tid := targetID
				notification := Notification{
					UserID:     receiverID,
					Type:       notifyType,
					ActorID:    &uid,
					TargetType: &tt,
					TargetID:   &tid,
					Content:    content,
				}
				_ = dao.DB.Create(&notification).Error
			}
		}

		incrementLikeCount(targetType, targetID)
		isLiked := true

		return &isLiked, nil
	}

	return nil, ErrInternal // 重试失败
}

// ToggleFavoriteService 切换收藏状态
func ToggleFavoriteService(uid uint, targetType uint8, targetID uint) (*bool, error) {
	const maxRetries = 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		var fav Favorite
		err := dao.DB.Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).
			First(&fav).Error

		if err == nil {
			// 已收藏 → 取消
			result := dao.DB.Where("user_id = ? AND target_type = ? AND target_id = ?", uid, targetType, targetID).
				Delete(&Favorite{})
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
		newFav := Favorite{
			UserID:     uid,
			TargetType: targetType,
			TargetID:   targetID,
		}
		if err := dao.DB.Create(&newFav).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				log.Printf("duplicate key on create favorite (attempt %d), retrying...", attempt)
				time.Sleep(50 * time.Millisecond)
				continue
			}
			log.Printf("create favorite failed (attempt %d): %v", attempt, err)
			return nil, ErrInternal
		}

		return &[]bool{true}[0], nil
	}

	return nil, ErrInternal // 重试失败
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

func GetFavoritesService(uid uint, page, size int) ([]FavoriteItem, int64, error) {
	offset := (page - 1) * size

	var total int64
	dao.DB.Model(&Favorite{}).
		Where("user_id = ?", uid).
		Count(&total)

	var favorites []Favorite
	err := dao.DB.Where("user_id = ?", uid).
		Offset(offset).
		Limit(size).
		Find(&favorites).Error

	if err != nil {
		return nil, 0, err
	}

	// 批量查收藏的帖子信息（从 posts 表）
	postIDs := make([]uint, 0, len(favorites))
	for _, f := range favorites {
		postIDs = append(postIDs, f.TargetID)
	}

	var posts []Post
	if len(postIDs) > 0 {
		dao.DB.Select("id, type, title, created_at").
			Where("id IN ? AND is_deleted = 0", postIDs).
			Find(&posts)
	}

	postMap := make(map[uint]Post)
	for _, p := range posts {
		postMap[p.ID] = p
	}

	items := make([]FavoriteItem, len(favorites))
	for i, f := range favorites {
		p, ok := postMap[f.TargetID]
		if ok {
			items[i] = FavoriteItem{
				ID:        p.ID,
				Type:      uint8(p.Type),
				Title:     p.Title,
				CreatedAt: p.CreatedAt,
			}
		} else {
			// 如果帖子不存在（已删除），可以跳过或返回空
			items[i] = FavoriteItem{ID: f.TargetID, Type: f.TargetType}
		}
	}

	return items, total, nil
}

func GetDraftService(uid uint, page, size int) ([]PostListItem, int64, error) {
	offset := (page - 1) * size

	// 1. 查询总数
	var total int64
	if err := dao.DB.Model(&Post{}).
		Where("author_id = ? AND status = 1 AND is_deleted = 0", uid).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 2. 查询草稿列表（一次查询就够了）
	var posts []Post
	err := dao.DB.Where("author_id = ? AND status = 1 AND is_deleted = 0", uid).
		Select("id, type, title, created_at, author_id, status"). // 明确指定需要的字段
		Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&posts).Error

	if err != nil {
		return nil, 0, err
	}

	// 3. 直接转换为返回结构
	items := make([]PostListItem, len(posts))
	for i, p := range posts {
		items[i] = PostListItem{
			ID:       p.ID,
			Type:     uint8(p.Type),
			Title:    p.Title,
			CreateAt: p.CreatedAt,
		}
	}

	return items, total, nil
}

func GetUnreadCountService(uid uint) (int64, error) {
	var count int64
	if err := dao.DB.Model(&Notification{}).
		Where("user_id = ? AND is_read = 0", uid).
		Count(&count).Error; err != nil {
		return 0, ErrInternal
	}
	return count, nil
}

func RefreshService(userID uint, tokenVersion int) (string, string, error) {
	var user User
	if err := dao.DB.Select("id", "username", "role", "token_version").
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		return "", "", ErrUnauthorized
	}

	if user.TokenVersion != tokenVersion {
		return "", "", ErrUnauthorized
	}

	res := dao.DB.Model(&User{}).
		Where("id = ? AND token_version = ?", userID, tokenVersion).
		Update("token_version", gorm.Expr("token_version + 1"))
	if res.Error != nil {
		return "", "", ErrInternal
	}
	if res.RowsAffected == 0 {
		return "", "", ErrUnauthorized
	}

	newVersion := tokenVersion + 1
	accessToken, err := GenerateToken(user.Username, user.ID, newVersion, user.Role)
	if err != nil {
		return "", "", ErrInternal
	}
	refreshToken, err := GenerateRefreshToken(user.ID, newVersion)
	if err != nil {
		return "", "", ErrInternal
	}

	return accessToken, refreshToken, nil
}

func MarkAllNotificationsRead(uid uint) error {
	if err := dao.DB.Model(&Notification{}).
		Where("user_id = ? AND is_read = 0", uid).
		Update("is_read", 1).Error; err != nil {
		return ErrInternal
	}
	return nil
}
