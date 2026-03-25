package main

import (
	"context"
	auth "example.com/micro-auth-demo/auth-service/kitex_gen/auth"
)

// AuthServiceImpl implements the last service interface defined in the IDL.
type AuthServiceImpl struct{}

// Login implements the AuthServiceImpl interface.
func (s *AuthServiceImpl) Login(ctx context.Context, req *auth.LoginRequest) (resp *auth.TokenPair, err error) {
	// TODO: Your code here...
	return
}

// RefreshToken implements the AuthServiceImpl interface.
func (s *AuthServiceImpl) RefreshToken(ctx context.Context, req *auth.RefreshTokenRequest) (resp *auth.TokenPair, err error) {
	// TODO: Your code here...
	return
}

// ValidateToken implements the AuthServiceImpl interface.
func (s *AuthServiceImpl) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (resp *auth.ValidateTokenResponse, err error) {
	// TODO: Your code here...
	return
}

// Logout implements the AuthServiceImpl interface.
func (s *AuthServiceImpl) Logout(ctx context.Context, req *auth.LogoutRequest) (resp *auth.CommonResponse, err error) {
	// TODO: Your code here...
	return
}

// LogoutAll implements the AuthServiceImpl interface.
func (s *AuthServiceImpl) LogoutAll(ctx context.Context, req *auth.LogoutAllRequest) (resp *auth.CommonResponse, err error) {
	// TODO: Your code here...
	return
}

// ListSessions implements the AuthServiceImpl interface.
func (s *AuthServiceImpl) ListSessions(ctx context.Context, req *auth.ListSessionsRequest) (resp *auth.ListSessionsResponse, err error) {
	// TODO: Your code here...
	return
}

// RevokeSession implements the AuthServiceImpl interface.
func (s *AuthServiceImpl) RevokeSession(ctx context.Context, req *auth.RevokeSessionRequest) (resp *auth.CommonResponse, err error) {
	// TODO: Your code here...
	return
}
