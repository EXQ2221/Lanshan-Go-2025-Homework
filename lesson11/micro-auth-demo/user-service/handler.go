package main

import (
	"context"
	user "example.com/micro-auth-demo/user-service/kitex_gen/user"
)

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct{}

// CreateUser implements the UserServiceImpl interface.
func (s *UserServiceImpl) CreateUser(ctx context.Context, req *user.CreateUserRequest) (resp *user.CreateUserResponse, err error) {
	// TODO: Your code here...
	return
}

// GetUser implements the UserServiceImpl interface.
func (s *UserServiceImpl) GetUser(ctx context.Context, req *user.GetUserRequest) (resp *user.GetUserResponse, err error) {
	// TODO: Your code here...
	return
}

// VerifyCredential implements the UserServiceImpl interface.
func (s *UserServiceImpl) VerifyCredential(ctx context.Context, req *user.VerifyCredentialRequest) (resp *user.VerifyCredentialResponse, err error) {
	// TODO: Your code here...
	return
}

// CheckPassword implements the UserServiceImpl interface.
func (s *UserServiceImpl) CheckPassword(ctx context.Context, req *user.CheckPasswordRequest) (resp *user.CheckPasswordResponse, err error) {
	// TODO: Your code here...
	return
}
