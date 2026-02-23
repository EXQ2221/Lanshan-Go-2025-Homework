package internal

import (
	"errors"
	"fmt"
	"lesson10/core"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func writeErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, core.ErrBadRequest):
		c.JSON(400, gin.H{"error": "bad_request"})
	case errors.Is(err, core.ErrUnauthorized):
		c.JSON(401, gin.H{"error": "unauthorized"})
	case errors.Is(err, core.ErrForbidden):
		c.JSON(403, gin.H{"error": "forbidden"})
	case errors.Is(err, core.ErrConflict):
		c.JSON(409, gin.H{"error": "conflict"})
	default:
		log.Println("internal error:", err)
		c.JSON(500, gin.H{"error": "server_error"})
	}
}
func RegisterHandler(c *gin.Context) {
	var req core.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "request format error",
		})
		return
	}

	user, err := core.RegisterService(req)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "register success",
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func LoginHandler(c *gin.Context) {
	var req core.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "req format error",
		})
		return
	}

	token, user, err := core.LoginService(req)
	if err != nil {

		errMsg := err.Error()

		switch {
		case strings.Contains(errMsg, "username incorrect"):
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "username incorrect",
			})
			return

		case strings.Contains(errMsg, "password incorrect"):
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "password incorrect",
			})
			return

		default:
			writeErr(c, err)
			return
		}
	}

	refreshToken, err := core.GenerateRefreshToken(user.ID, user.TokenVersion)
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":       "login success",
		"user_id":       user.ID,
		"username":      user.Username,
		"token":         token,
		"refresh_token": refreshToken,
	})

}

func ChangePassHandler(c *gin.Context) {
	var req core.ChangePassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "req format error",
		})
		return
	}

	id := c.GetUint("user_id")

	err := core.ChangePassService(req, id)

	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"need relog in": true,
	})
}

func UpdateProfileHandler(c *gin.Context) {
	var req core.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "req format error",
		})
		return
	}

	id := c.GetUint("user_id")

	err := core.UpdateProfileService(req, id)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok": true,
	})
}

func UploadAvatarHandler(c *gin.Context) {
	userID := c.GetUint("user_id")

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(400, gin.H{"error": "missing avatar file"})
		return
	}

	// 1) 大小限制
	const maxSize = 5 * 1024 * 1024
	if file.Size > maxSize {
		c.JSON(400, gin.H{"error": "file too large (max 5MB)"})
		return
	}

	// 2) 打开读头部，检查 mime（防止随便传 .exe）
	f, err := file.Open()
	if err != nil {
		c.JSON(500, gin.H{"error": "open file failed"})
		return
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		c.JSON(400, gin.H{"error": "only jpg/png/webp allowed"})
		return
	}

	// 3) 生成文件名（避免重名/路径穿越）
	ext := ".jpg"
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	}

	// 你也可以用 uuid，这里用 时间戳+userID
	filename := fmt.Sprintf("u%d_%d%s", userID, time.Now().UnixNano(), ext)

	// 4) 确保目录存在
	saveDir := "static/uploads/avatars"
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		c.JSON(500, gin.H{"error": "mkdir failed"})
		return
	}

	savePath := filepath.Join(saveDir, filename)

	// 注意：刚刚读了512字节，不影响 SaveUploadedFile（它会重新打开文件）
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(500, gin.H{"error": "save file failed"})
		return
	}

	// 5) 写库：avatar_url 存一个可访问的 url
	avatarURL := "/static/uploads/avatars/" + filename
	if err := core.UpdateAvatarService(userID, avatarURL); err != nil {
		c.JSON(500, gin.H{"error": "db update failed"})
		return
	}

	c.JSON(200, gin.H{"avatar_url": avatarURL})
}

func CreatePostHandler(c *gin.Context) {
	var req core.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "req format error",
		})
		return
	}

	authorID := c.GetUint("user_id")
	if authorID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "fail to find user id",
		})
		return
	}

	if req.Status == 0 {
		req.Status = 0 //发布状态
	}
	post, err := core.CreatePostService(&req, authorID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":   true,
		"post": post,
	})
}

func ListPostsHandler(c *gin.Context) {
	var q core.ListPostsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query format error"})
		return
	}

	list, total, err := core.ListPostsService(q)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"list":  list,
		"total": total,
		"page": func() int {
			if q.Page == 0 {
				return 1
			}
			return q.Page
		}(),
		"page_size": func() int {
			if q.PageSize == 0 {
				return 20
			}
			return q.PageSize
		}(),
	})
}

func GetPostHandler(c *gin.Context) {
	postID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || postID64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id format error"})
		return
	}

	currentUserID := c.GetUint("user_id")

	resp, err := core.GetPostService(currentUserID, uint(postID64))
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func PostCommentHandler(c *gin.Context) {
	var req core.PostCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "format error",
		})
		return
	}

	id := c.GetUint("user_id")

	comment, err := core.PostCommentService(id, &req)
	if err != nil {
		errMsg := err.Error()

		switch {
		case strings.Contains(errMsg, "fail to find comment"):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "fail to find comment",
			})
			return

		case strings.Contains(errMsg, "fail to find post or question"):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "fail to find post or question",
			})
			return

		default:
			writeErr(c, err)
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "post success",
		"comment": comment,
	})
}

func GetCommentsHandler(c *gin.Context) {
	var req core.GetCommentsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "format error",
		})
		return
	}

	resp, err := core.GetCommentsService(&req)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    resp,
	})
}

func GetRepliesHandler(c *gin.Context) {
	parentIDStr := c.Param("parent_id")
	parentID, err := strconv.ParseUint(parentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid parent id"})
		return
	}

	uid := c.GetUint("user_id")

	replies, total, err := core.GetAllReplies(uint(parentID), uid)
	if err != nil {
		log.Printf("get replies failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"replies": replies,
			"total":   total,
		},
	})
}
func UpdatePostHandler(c *gin.Context) {
	PostIDString := c.Param("id")
	PostID, err := strconv.ParseUint(PostIDString, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id format incorrect",
		})
		return
	}

	var req core.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "format error",
		})
		return
	}

	id := c.GetUint("user_id")
	if id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "please log in",
		})
		return
	}

	err = core.UpdatePostService(PostID, id, &req)

	if err != nil {
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "fail to find post"):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "fail to find post",
			})
			return
		case strings.Contains(errMsg, "unauthorized"):
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "unauthorized",
			})
			return
		default:
			log.Printf("error: %v", err)
			writeErr(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "update success",
	})
}

func DeletePostHandler(c *gin.Context) {
	postIDStr := c.Param("id")
	postID, err := strconv.ParseUint(postIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "id format incorrect",
		})
		return
	}

	uid := c.GetUint("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	role := c.GetUint("role")
	err = core.DeletePostService(uint(postID), uid, role)
	if err != nil {
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "fail to find post"):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "fail to find post",
			})
			return

		default:
			log.Printf("error: %v", err)
			writeErr(c, err)
			return
		}

	}

	c.JSON(http.StatusOK, gin.H{
		"message": "delete success",
	})
}

func DeleteCommentHandler(c *gin.Context) {
	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseUint(commentIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id format incorrect",
		})
		return
	}

	uid := c.GetUint("user_id")
	role := c.GetUint("role")

	err = core.DeleteComment(uint(commentID), uid, role)
	if err != nil {
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "fail to find comment"):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "fail to find comment",
			})
			return

		default:
			log.Printf("error: %v", err)
			writeErr(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func GetUserInfoHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	userIDUint64, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id format error",
		})
		return
	}

	page := 1
	pageStr := c.DefaultQuery("page", "1")
	if p, err := strconv.Atoi(pageStr); err == nil && p >= 1 {
		page = p
	}

	currentID := c.GetUint("user_id")
	userPublicInfo, err := core.GetUserInfoService(currentID, uint(userIDUint64), page)
	if err != nil {
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "fail to find user"):
			c.JSON(http.StatusNotFound, gin.H{
				"message": "fail to find user",
			})
			return

		default:
			log.Printf("error: %v", err)
			writeErr(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    userPublicInfo,
	})
}

func FollowUserHandler(c *gin.Context) {
	followeeIDStr := c.Param("id")
	followeeID, err := strconv.ParseUint(followeeIDStr, 10, 64)
	if err != nil || followeeID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id format incorrect",
		})
		return
	}

	followerID := c.GetUint("user_id")
	if followerID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "please log in",
		})
		return
	}

	if followerID == uint(followeeID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "can not follow yourself",
		})
		return
	}

	err = core.FollowUserService(followerID, uint(followeeID))

	if err != nil {
		errMsg := err.Error()
		switch {
		case strings.Contains(errMsg, "has followed"):
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "has followed",
			})
			return
		default:
			log.Printf("fail: %v", err)
			writeErr(c, err)
			return
		}

	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func UnfollowUserHandler(c *gin.Context) {
	followeeIDStr := c.Param("id")
	followeeID, err := strconv.ParseUint(followeeIDStr, 10, 64)

	if err != nil || followeeID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id format incorrect",
		})
	}

	followerID := c.GetUint("user_id")
	if followerID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "please log in",
		})
	}

	err = core.UnfollowUserService(followerID, uint(followeeID))
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "has not followed") {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "has not followed",
			})
			return

		} else {
			writeErr(c, err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func UploadArticleImageHandler(c *gin.Context) {
	uid := c.GetUint("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "please log in",
		})
		return
	}

	file, err := c.FormFile("image") // 前端 form-data 字段名统一用 "image"
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "please upload image",
		})
		return
	}

	// 1. 大小限制：5MB
	const maxSize = 10 * 1024 * 1024
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "max 10MB"})
		return
	}

	// 2. 打开文件读头部，检测真实 MIME 类型（防伪造）
	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail to open file"})
		return
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	contentType := http.DetectContentType(buf[:n])

	allowedTypes := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/webp": ".webp",
	}

	ext, ok := allowedTypes[contentType]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "only jpg/png/webp",
		})
		return
	}

	// 3. 生成唯一文件名
	filename := fmt.Sprintf("article_%d_%d%s", uid, time.Now().UnixNano(), ext)

	// 4. 确保目录存在
	saveDir := "static/uploads/images"
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail to make dir",
		})
		return
	}

	savePath := filepath.Join(saveDir, filename)

	// 5. 保存文件（Gin 内置方法会重新打开文件，所以前面读 512 字节不影响）
	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "fail to save file",
		})
		return
	}

	// 6. 构造可访问的 URL（相对路径，配合 r.Static 使用）
	imageURL := "/static/uploads/images/" + filename

	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"image_url": imageURL,
	})
}

// 点赞
func ToggleReactionHandler(c *gin.Context) {
	var req core.LikeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "format error",
		})
		return
	}

	uid := c.GetUint("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "please login",
		})
		return
	}

	isLiked, err := core.ToggleReactionService(uid, req.TargetType, req.TargetID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"status":  isLiked,
	})
}

// 收藏
func ToggleFavoriteHandler(c *gin.Context) {
	var req core.FavorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "format error",
		})
		return
	}

	uid := c.GetUint("user_id")
	if uid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "please login",
		})
		return
	}

	isFavorited, err := core.ToggleFavoriteService(uid, req.TargetType, req.TargetID)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    gin.H{"is_favorited": *isFavorited}, // 操作后是否已收藏

	})
}

func GetFollowersHandler(c *gin.Context) {
	getFollowList(c, "followers")
}

func GetFollowingHandler(c *gin.Context) {
	getFollowList(c, "following")
}

func getFollowList(c *gin.Context, listType string) {
	targetUserIDStr := c.Param("id")
	targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 64)
	if err != nil || targetUserID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user id"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	sizeStr := c.DefaultQuery("size", "20")
	size, _ := strconv.Atoi(sizeStr)
	if size < 1 || size > 50 {
		size = 20
	}

	currentUserID := c.GetUint("user_id") // 当前登录用户（用于 is_followed，可选）

	users, total, err := core.GetFollowListService(uint(targetUserID), listType, currentUserID, page, size)
	if err != nil {
		log.Printf("get follow list failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"users": users,
			"total": total,
			"page":  page,
			"size":  size,
		},
	})
}

func GetNotificationsHandler(c *gin.Context) {
	uid := c.GetUint("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "please login first",
		})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	sizeStr := c.DefaultQuery("size", "20")
	size, _ := strconv.Atoi(sizeStr)
	if size < 1 || size > 50 {
		size = 20
	}

	unreadOnly := c.DefaultQuery("unread_only", "0") == "1"

	notifications, total, err := core.GetNotifications(uid, page, size, unreadOnly)
	if err != nil {
		log.Printf("get notifications failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"notifications": notifications,
			"total":         total,
			"page":          page,
			"size":          size,
		},
	})
}

func GetFavoritesHandler(c *gin.Context) {
	uid := c.GetUint("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "please login first"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if size < 1 || size > 50 {
		size = 20
	}

	favorites, total, err := core.GetFavoritesService(uid, page, size)
	if err != nil {
		log.Printf("get favorites failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"favorites": favorites,
			"total":     total,
			"page":      page,
			"size":      size,
		},
	})
}

func GetDraftHandler(c *gin.Context) {
	uid := c.GetUint("user_id")
	if uid == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"message": "please login first"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if size < 1 || size > 50 {
		size = 20
	}

	drafts, total, err := core.GetDraftService(uid, page, size)
	if err != nil {
		log.Printf("get favorites failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data": gin.H{
			"drafts": drafts,
			"total":  total,
			"page":   page,
			"size":   size,
		},
	})
}

func GetUnreadCountHandler(c *gin.Context) {
	uid := c.GetUint("user_id")

	count, err := core.GetUnreadCountService(uid)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(200, gin.H{"count": count})
}

func RefreshHandler(c *gin.Context) {
	var req core.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(400, gin.H{
			"error": "invalid refresh token",
		})
		return
	}

	token, err := core.ValidateToken(req.RefreshToken)

	if err != nil || !token.Valid {
		c.JSON(401, gin.H{"error": "invalid refresh token"})
		return
	}
	claims := token.Claims.(jwt.MapClaims)

	if claims["type"] != "refresh" {
		c.JSON(401, gin.H{"error": "invalid token type"})
		return
	}
	userID := uint(claims["user_id"].(float64))
	tokenVersion := int(claims["token_version"].(float64))

	newAccessToken, newRefreshToken, err := core.RefreshService(userID, tokenVersion)
	if err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(200, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})

}

func MarkAllNotificationsReadHandler(c *gin.Context) {
	uid := c.GetUint("user_id")
	if uid == 0 {
		c.JSON(401, gin.H{"message": "please login first"})
		return
	}

	if err := core.MarkAllNotificationsRead(uid); err != nil {
		writeErr(c, err)
		return
	}

	c.JSON(200, gin.H{"message": "success"})
}
