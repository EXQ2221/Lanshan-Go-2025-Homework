package handler

import (
	"fmt"
	"lesson10/internal/dto"
	"lesson10/internal/pkg/response"
	"lesson10/internal/service"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CreatePostHandler(postSvc *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.CreatePostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "req format error")
			return
		}

		authorID := c.GetUint("user_id")
		if authorID == 0 {
			response.Error(c, http.StatusUnauthorized, "fail to find user id")
			return
		}

		if req.Status == 0 {
			req.Status = 0 //发布状态
		}
		post, err := postSvc.CreatePostService(c.Request.Context(), &req, authorID)
		if err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{
			"ok":   true,
			"post": post,
		})
	}
}

func ListPostsHandler(postSvc *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var q dto.ListPostsQuery
		if err := c.ShouldBindQuery(&q); err != nil {
			response.Error(c, http.StatusBadRequest, "query format error")
			return
		}

		list, total, err := postSvc.ListPostsService(c.Request.Context(), q)
		if err != nil {
			writeErr(c, err)
			return
		}

		response.OK(c, gin.H{
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
}

func GetPostHandler(postSvc *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || postID64 == 0 {
			response.Error(c, http.StatusBadRequest, "id format error")
			return
		}

		currentUserID := c.GetUint("user_id")

		resp, err := postSvc.GetPostService(c.Request.Context(), currentUserID, uint(postID64))
		if err != nil {
			writeErr(c, err)
			return
		}
		response.OK(c, resp)
	}
}

func UpdatePostHandler(postSvc *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		PostIDString := c.Param("id")
		PostID, err := strconv.ParseUint(PostIDString, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "id format incorrect")
			return
		}

		var req dto.UpdatePostRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, "format error")
			return
		}

		id := c.GetUint("user_id")
		if id == 0 {
			response.Error(c, http.StatusUnauthorized, "please log in")
			return
		}

		err = postSvc.UpdatePostService(c.Request.Context(), PostID, id, &req)

		if err != nil {
			writeErr(c, err)
			return
		}
		response.OK(c, gin.H{
			"ok":      true,
			"message": "update success",
		})
	}
}

func DeletePostHandler(postSvc *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postIDStr := c.Param("id")
		postID, err := strconv.ParseUint(postIDStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "id format incorrect")
			return
		}

		uid := c.GetUint("user_id")
		if uid == 0 {
			response.Error(c, http.StatusUnauthorized, "unauthorized")
			return
		}

		role := c.GetUint("role")
		err = postSvc.DeletePostService(c.Request.Context(), uint(postID), uid, role)
		if err != nil {
			writeErr(c, err)
			return
		}
		response.OK(c, gin.H{
			"message": "delete success",
		})
	}
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

func GetFavoritesHandler(postSvc *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		favorites, total, err := postSvc.GetFavoritesService(c.Request.Context(), uid, page, size)
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
}

func GetDraftHandler(postSvc *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		drafts, total, err := postSvc.GetDraftService(c.Request.Context(), uid, page, size)
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
}
