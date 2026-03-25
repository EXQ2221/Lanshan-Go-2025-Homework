package handler

import (
	"context"

	"example.com/micro-auth-demo/user-service/internal/biz"
	"example.com/micro-auth-demo/user-service/internal/pkg/convert"
	userpb "example.com/micro-auth-demo/user-service/kitex_gen/user"
)

type UserServiceImpl struct {
	Service *biz.UserService
}

func NewUserServiceImpl(service *biz.UserService) *UserServiceImpl {
	return &UserServiceImpl{Service: service}
}

func (h *UserServiceImpl) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	user, err := h.Service.CreateUser(ctx, req.Email, req.Nickname, req.Password)
	if err != nil {
		return nil, err
	}

	return &userpb.CreateUserResponse{
		User: convert.ToUserInfo(user),
	}, nil
}

func (h *UserServiceImpl) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	user, err := h.Service.GetByID(ctx, req.UserId)
	if err != nil {
		return nil, err
	}

	return &userpb.GetUserResponse{
		User: convert.ToUserInfo(user),
	}, nil
}

func (h *UserServiceImpl) VerifyCredential(ctx context.Context, req *userpb.VerifyCredentialRequest) (*userpb.VerifyCredentialResponse, error) {
	user, ok, err := h.Service.VerifyCredential(ctx, req.Email, req.Password)
	if err != nil {
		return nil, err
	}

	resp := &userpb.VerifyCredentialResponse{
		Ok: ok,
	}
	if ok {
		resp.User = convert.ToUserInfo(user)
	} else {
		resp.Reason = "invalid email or password"
	}

	return resp, nil
}

func (h *UserServiceImpl) CheckPassword(ctx context.Context, req *userpb.CheckPasswordRequest) (*userpb.CheckPasswordResponse, error) {
	ok, err := h.Service.CheckPassword(ctx, req.UserId, req.Password)
	if err != nil {
		return nil, err
	}

	return &userpb.CheckPasswordResponse{Ok: ok}, nil
}
