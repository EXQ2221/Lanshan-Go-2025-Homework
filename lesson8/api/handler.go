package api

import (
	"lesson8/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateTodoHandler(c *gin.Context) {
	var todo model.Todo
	if err := c.ShouldBindJSON(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "format incorrect",
		})
		return
	}

	userID := c.GetUint("user_id")
	err := CreateTodoService(userID, &todo)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "create successfully",
		"title":   todo.Title,
		"user_id": todo.UserID,
	})
}

func GetTodosHandler(c *gin.Context) {
	userID := c.GetUint("user_id")
	todos, err := GetTodosService(userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"todos": todos,
	})
}

func RegisterHandler(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "request format error",
		})
		return
	}

	user, err := RegisterService(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "register success",
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func LoginHandler(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "format error",
		})
		return
	}

	token, user, err := LoginService(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
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

func UpdateTodoHandler(c *gin.Context) {
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

	updatedTodo, err := UpdateTodoService(req, todoID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "update success",
		"todo":    updatedTodo,
		"user_id": userID,
	})
}

func DeleteTodoHandler(c *gin.Context) {
	todoID := c.Param("id")
	if todoID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "can not find todo id",
		})
		return
	}

	userID := c.GetUint("user_id")
	err := DeleteTodoService(todoID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "delete success",
		"todo_id": todoID,
		"user_id": userID,
	})
}
