package handler

import (
	"context"

	"example.com/micro-auth-demo/auth-service/internal/biz"
	"example.com/micro-auth-demo/auth-service/internal/pkg/convert"
	authpb "example.com/micro-auth-demo/auth-service/kitex_gen/auth"
)

type AuthServiceImpl struct {
	Service *biz.AuthService
}

func NewAuthServiceImpl(service *biz.AuthService) *AuthServiceImpl {
	return &AuthServiceImpl{Service: service}
}

func (h *AuthServiceImpl) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.TokenPair, error) {
	pair, err := h.Service.Login(ctx, biz.LoginInput{
		Email:      req.Email,
		Password:   req.Password,
		DeviceID:   req.DeviceId,
		DeviceName: req.DeviceName,
		UserAgent:  req.UserAgent,
		IP:         req.Ip,
	})
	if err != nil {
		return nil, err
	}

	return convert.ToTokenPair(pair), nil
}

func (h *AuthServiceImpl) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.TokenPair, error) {
	pair, err := h.Service.Refresh(ctx, biz.RefreshInput{
		RefreshToken: req.RefreshToken,
		DeviceID:     req.DeviceId,
		UserAgent:    req.UserAgent,
		IP:           req.Ip,
	})
	if err != nil {
		return nil, err
	}

	return convert.ToTokenPair(pair), nil
}

func (h *AuthServiceImpl) ValidateToken(ctx context.Context, req *authpb.ValidateTokenRequest) (*authpb.ValidateTokenResponse, error) {
	identity, err := h.Service.ValidateToken(ctx, req.AccessToken)
	if err != nil {
		return &authpb.ValidateTokenResponse{
			Valid:  false,
			Reason: err.Error(),
		}, nil
	}

	return &authpb.ValidateTokenResponse{
		Valid:     true,
		UserId:    identity.UserID,
		SessionId: identity.SessionID,
	}, nil
}

func (h *AuthServiceImpl) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.CommonResponse, error) {
	if err := h.Service.Logout(ctx, req.AccessToken); err != nil {
		return &authpb.CommonResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &authpb.CommonResponse{Success: true, Message: "ok"}, nil
}

func (h *AuthServiceImpl) LogoutAll(ctx context.Context, req *authpb.LogoutAllRequest) (*authpb.CommonResponse, error) {
	if err := h.Service.LogoutAll(ctx, req.UserId, req.Password); err != nil {
		return &authpb.CommonResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &authpb.CommonResponse{Success: true, Message: "ok"}, nil
}

func (h *AuthServiceImpl) ListSessions(ctx context.Context, req *authpb.ListSessionsRequest) (*authpb.ListSessionsResponse, error) {
	sessions, err := h.Service.ListSessions(ctx, req.UserId, req.CurrentSessionId)
	if err != nil {
		return nil, err
	}

	return &authpb.ListSessionsResponse{
		Sessions: convert.ToSessionInfos(sessions),
	}, nil
}

func (h *AuthServiceImpl) RevokeSession(ctx context.Context, req *authpb.RevokeSessionRequest) (*authpb.CommonResponse, error) {
	if err := h.Service.RevokeSession(ctx, req.UserId, req.SessionId, req.Password); err != nil {
		return &authpb.CommonResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &authpb.CommonResponse{Success: true, Message: "ok"}, nil
}
