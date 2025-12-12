package api

import (
	"lesson7/dao"
	"lesson7/model"
	"lesson7/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong",
		})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "encryption error",
		})
		return
	}

	user := model.User{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if err := dao.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusConflict, gin.H{
			"message": "username has exist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "register success",
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "format error",
		})
		return
	}

	var user model.User
	if err := dao.DB.Where("username = ? ", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "username or password incorrect",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "password incorrect",
		})
		return
	}

	token, err := utils.GenerateToken(user.Username, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "fail to generate jwt",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "login success",
		"user_id":  user.ID,
		"username": user.Username,
		"token":    token,
	})
}

func CreateTodo(c *gin.Context) {
	var todo model.Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "format incorrect",
		})
		return
	}

	userID := c.GetUint("user_id")
	todo.UserID = userID

	if err := dao.DB.Create(&todo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "fail to create todo list",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "create successfully",
		"title":   todo.Title,
		"user_id": todo.UserID,
	})
}

func GetTodos(c *gin.Context) {
	var todos []model.Todo
	userID := c.GetUint("user_id")
	if err := dao.DB.Where("user_id = ?", userID).Find(&todos).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "fail to get todo list",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"todos": todos,
	})
}

func UpdateTodo(c *gin.Context) {
	var req model.UpdateDataRequest
	todoID := c.Param("id")
	if todoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "todo id not exist",
		})
		return
	}

	userID := c.GetUint("user_id")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "format incorrect",
		})
		return
	}

	if req.Title == nil && req.Done == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "please update at least one status",
		})
		return
	}

	updates := make(map[string]interface{})

	if req.Title != nil {
		updates["title"] = req.Title
	}
	if req.Done != nil {
		updates["done"] = req.Done
	}

	result := dao.DB.Model(&model.Todo{}).Where("id = ? AND user_id = ?", todoID, userID).Updates(updates)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "fail to revise",
		})
		return
	}

	var updatedTodo model.Todo
	dao.DB.Where("id = ?", todoID).First(&updatedTodo)

	c.JSON(http.StatusOK, gin.H{
		"message": "update success",
		"todo":    updatedTodo,
		"user_id": userID,
	})
}

func DeleteTodo(c *gin.Context) {
	todoID := c.Param("id")
	if todoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "can not find todo id",
		})
		return
	}

	userID := c.GetUint("user_id")

	result := dao.DB.Where("id = ? AND user_id = ?", todoID, userID).Delete(&model.Todo{})
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "fail to delete todo ",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "delete success",
		"todo_id": todoID,
		"user_id": userID,
	})
}
