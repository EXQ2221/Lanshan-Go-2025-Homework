package handler

import (
	"lesson10/internal/dto"
	"lesson10/internal/service"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func PostCommentHandler(commentSvc *service.CommentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.PostCommentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "format error",
			})
			return
		}

		id := c.GetUint("user_id")

		comment, err := commentSvc.PostCommentService(c.Request.Context(), id, &req)
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
}

func GetCommentsHandler(commentSvc *service.CommentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req dto.GetCommentsReq
		if err := c.ShouldBindQuery(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "format error",
			})
			return
		}

		resp, err := commentSvc.GetCommentsService(c.Request.Context(), &req)
		if err != nil {
			writeErr(c, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"data":    resp,
		})
	}
}

func GetRepliesHandler(commentSvc *service.CommentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		parentIDStr := c.Param("parent_id")
		parentID, err := strconv.ParseUint(parentIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid parent id"})
			return
		}

		uid := c.GetUint("user_id")

		replies, total, err := commentSvc.GetAllReplies(c.Request.Context(), uint(parentID), uid)
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
}

func DeleteCommentHandler(commentSvc *service.CommentService) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		err = commentSvc.DeleteComment(c.Request.Context(), uint(commentID), uid, role)
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
}
