package convert

import (
	"example.com/micro-auth-demo/user-service/internal/dal/model"
	userpb "example.com/micro-auth-demo/user-service/kitex_gen/user"
)

func ToUserInfo(user *model.User) *userpb.UserInfo {
	if user == nil {
		return nil
	}

	return &userpb.UserInfo{
		UserId:   user.ID,
		Email:    user.Email,
		Nickname: user.Nickname,
	}
}
