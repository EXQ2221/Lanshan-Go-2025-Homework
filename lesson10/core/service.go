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
		Where("is_deleted = 0")


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

	return &PostDetailResp{
		ID:         p.ID,
		Type:       uint8(p.Type),
		AuthorID:   p.AuthorID,
		AuthorName: p.Author.Username,
		Title:      p.Title,
		Content:    p.Content,
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

		if err == gorm.ErrRecordNotFound {
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

	if err := dao.DB.Create(comment).Error; err != nil {
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
	query := dao.DB.Where("target_type = ? AND target_id = ? AND depth = 0 AND is_deleted = 0",
		req.TargetType, req.TargetID)

	var total int64
	query.Count(&total)

	// 分页 排序
	err := query.Order("created_at DESC").
		Offset(offset).
		Limit(req.Size).
		Find(&comments).Error

	if err != nil {
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

func UpdatePostService(PostID uint64, id uint, req *UpdatePostRequest) error {
	var post Post
	err := dao.DB.Where("id = ? AND is_deleted = 0", PostID).First(&post).Error
	if err == gorm.ErrRecordNotFound {
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
	if err == gorm.ErrRecordNotFound {
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
