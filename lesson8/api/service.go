package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"lesson8/dao"
	"lesson8/model"
	"lesson8/utils"
	"log"
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func CreateTodoService(userID uint, todo *model.Todo) error {
	todo.UserID = userID

	if err := dao.DB.Create(&todo).Error; err != nil {
		return err
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:todos", userID)
	
	dao.Redis.Del(ctx, cacheKey)

	return nil
}

func GetTodosService(userID uint) ([]model.Todo, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:todos", userID)
	todoMap, err := dao.Redis.HGetAll(ctx, cacheKey).Result()
	ttl := 10*time.Minute + time.Duration(rand.Intn(60))*time.Second

	if err != nil {
		log.Printf("redis get failed, key=%s, err=%v", cacheKey, err)
	} else if len(todoMap) > 0 {
		var todos []model.Todo

		if _, ok := todoMap["empty"]; ok {
			return []model.Todo{}, nil
		}
		for _, todoJSON := range todoMap {

			var todo model.Todo
			if err := json.Unmarshal([]byte(todoJSON), &todo); err == nil {
				todos = append(todos, todo)
			}
		}
		return todos, nil
	}
	var todos []model.Todo
	if err := dao.DB.Where("user_id = ?", userID).Find(&todos).Error; err != nil {
		log.Printf("db falied, err=%v", err)
		return nil, err
	}

	if len(todos) == 0 {
		dao.Redis.HSet(ctx, cacheKey, "empty", "1")
		dao.Redis.Expire(ctx, cacheKey, ttl)
		return []model.Todo{}, nil
	}

	for _, todo := range todos {
		todoJSON, _ := json.Marshal(todo)
		field := fmt.Sprintf("%d", todo.ID)
		dao.Redis.HSet(ctx, cacheKey, field, todoJSON)
	}

	dao.Redis.Expire(ctx, cacheKey, ttl)
	return todos, nil
}

func RegisterService(req model.RegisterRequest) (*model.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Username: req.Username,
		Password: string(hashedPassword),
	}

	if err := dao.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func LoginService(req model.LoginRequest) (string, *model.User, error) {
	var user model.User
	if err := dao.DB.Where("username = ? ", req.Username).First(&user).Error; err != nil {
		return "", nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return "", nil, err
	}

	token, err := utils.GenerateToken(user.Username, user.ID)
	if err != nil {
		return "", &user, err
	}

	return token, &user, nil
}

func UpdateTodoService(req model.UpdateDataRequest, todoID string, userID uint) (*model.Todo, error) {

	if req.Title == nil && req.Done == nil {
		return nil, errors.New("no field to update")
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
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("todo not found")
	}

	var updatedTodo model.Todo
	dao.DB.Where("id = ?", todoID).First(&updatedTodo)

	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:todos", userID)

	field := todoID

	if currentTodo, err := json.Marshal(updatedTodo); err == nil {
		dao.Redis.HSet(ctx, cacheKey, field, currentTodo)
	}
	return &updatedTodo, nil
}

func DeleteTodoService(todoID string, userID uint) error {
	result := dao.DB.Where("id = ? AND user_id = ?", todoID, userID).Delete(&model.Todo{})
	if result.Error != nil {
		return result.Error
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("user:%d:todos", userID)
	field := todoID
	dao.Redis.HDel(ctx, cacheKey, field)

	return nil
}
