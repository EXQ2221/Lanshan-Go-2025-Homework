package rpc

import (
	"context"

	userpb "example.com/micro-auth-demo/auth-service/kitex_gen/user"
	"example.com/micro-auth-demo/auth-service/kitex_gen/user/userservice"
	"github.com/cloudwego/kitex/client"
)

type UserInfo struct {
	UserID   int64
	Email    string
	Nickname string
}

type UserClient interface {
	VerifyCredential(ctx context.Context, email, password string) (*UserInfo, bool, error)
	CheckPassword(ctx context.Context, userID int64, password string) (bool, error)
}

type KitexUserClient struct {
	client userservice.Client
}

func NewUserClient(addr string) (*KitexUserClient, error) {
	c, err := userservice.NewClient("user-service", client.WithHostPorts(addr))
	if err != nil {
		return nil, err
	}
	return &KitexUserClient{client: c}, nil
}

func (c *KitexUserClient) VerifyCredential(ctx context.Context, email, password string) (*UserInfo, bool, error) {
	resp, err := c.client.VerifyCredential(ctx, &userpb.VerifyCredentialRequest{
		Email:    email,
		Password: password,
	})
	if err != nil {
		return nil, false, err
	}
	if !resp.Ok || resp.User == nil {
		return nil, false, nil
	}

	return &UserInfo{
		UserID:   resp.User.UserId,
		Email:    resp.User.Email,
		Nickname: resp.User.Nickname,
	}, true, nil
}

func (c *KitexUserClient) CheckPassword(ctx context.Context, userID int64, password string) (bool, error) {
	resp, err := c.client.CheckPassword(ctx, &userpb.CheckPasswordRequest{
		UserId:   userID,
		Password: password,
	})
	if err != nil {
		return false, err
	}
	return resp.Ok, nil
}
