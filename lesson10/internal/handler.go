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
	c.JSON(http.StatusOK, gin.H{
		"message":  "login success",
		"user_id":  user.ID,
		"username": user.Username,
		"token":    token,
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
		"ok":           true,
		"need relogin": true,
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
	saveDir := "uploads/avatars"
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
	avatarURL := "/static/avatars/" + filename
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
	id64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id format error"})
		return
	}

	resp, err := core.GetPostService(uint(id64))
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

	c.JSON(200, gin.H{
		"code":    200,
		"message": "success",
		"data":    resp,
	})
}

func UpdatePostHandler(c *gin.Context) {
	PostIDString := c.Param("id")
	PostID, err := strconv.ParseUint(PostIDString, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "id format incorrect",
		})
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
		case strings.Contains(errMsg, "unauthorized"):
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "unauthorized",
			})
		default:
			log.Printf("error: %v", err)
			writeErr(c, err)
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

		default:
			log.Printf("error: %v", err)
			writeErr(c, err)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "delete success",
	})
}
