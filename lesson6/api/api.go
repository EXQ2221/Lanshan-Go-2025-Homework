package api

import (
	"lesson6/dao"
	"lesson6/model"
	"lesson6/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Login(c *gin.Context) {
	var req model.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "bad request",
		})
		return
	}
	if !dao.CheckUser(req.Username, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "username or password incorrect",
		})
		return
	}

	tokenString, err := utils.GenerateToken(req.Username)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "fail to generate toekn",
		})
		return
	}

	refreshTokenString, err := utils.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "fail to generate toekn",
		})
		return
	}

	dao.AddRefreshToken(refreshTokenString, req.Username)

	c.JSON(200, gin.H{
		"message":       "login seccess",
		"token":         tokenString,
		"refresh_token": refreshTokenString,
	})
}

func Register(c *gin.Context) {
	var req model.User

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "bad request",
		})
		return
	}

	if _, ok := dao.GetPassword(req.Username); ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "user already exists",
		})
		return
	}

	dao.AddUser(req.Username, req.Password)
	err := dao.SaveDB()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "fail to save data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "register success",
	})
}

func ChangePassword(c *gin.Context) {
	var req model.Changepw
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "bad request",
		})
		return
	}

	if !dao.UserExist(req.Username) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "username not exist",
		})
		return
	}

	if pwg, ok := dao.GetPassword(req.Username); ok {
		if pwg == req.OldPass {
			dao.AddUser(req.Username, req.Newpass)
			err := dao.SaveDB()
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "fail to save data",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "change password success",
			})
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "old password incorrect",
			})
		}
	}
}

func Refresh(c *gin.Context) {
	var req model.RefreshRequest

	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "fail to find refresh token",
		})
	}

	username, ok := dao.ValidateRefreshToken(req.RefreshToken)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid or expired refresh token",
		})
		return
	}

	newjwt, err := utils.GenerateToken(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "fail to generate new jwt",
		})
		return
	}

	newRefreshToken, err := utils.GenerateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "fail to generate nwe fresh token",
		})
		return
	}

	dao.AddRefreshToken(newRefreshToken, username)

	c.JSON(http.StatusOK, gin.H{
		"token":         newjwt,
		"refresh_token": newRefreshToken,
	})
}
