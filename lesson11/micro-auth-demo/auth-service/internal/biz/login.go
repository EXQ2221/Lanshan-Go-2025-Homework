package biz

import (
	"context"
	"time"

	"example.com/micro-auth-demo/auth-service/internal/dal/model"
	browserprofile "example.com/micro-auth-demo/auth-service/internal/pkg/browser"
	"example.com/micro-auth-demo/auth-service/internal/pkg/token"
	"gorm.io/gorm"
)

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*TokenPair, error) {
	userInfo, ok, err := s.UserClient.VerifyCredential(ctx, input.Email, input.Password)
	if err != nil {
		return nil, err
	}
	if !ok || userInfo == nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now()
	sessionID, err := token.Generate(16)
	if err != nil {
		return nil, err
	}
	refreshTokenValue, err := NewRefreshToken()
	if err != nil {
		return nil, err
	}
	accessToken, accessJTI, accessExpiresAt, err := s.issueAccessToken(userInfo.UserID, sessionID, now)
	if err != nil {
		return nil, err
	}

	refreshExpiresAt := now.Add(s.RefreshTTL)
	browserInfo := browserprofile.Parse(input.UserAgent)
	session := &model.Session{
		SessionID:            sessionID,
		UserID:               userInfo.UserID,
		Status:               SessionStatusActive,
		DeviceID:             coalesce(input.DeviceID, sessionID),
		DeviceName:           coalesce(input.DeviceName, "unknown-device"),
		UserAgent:            input.UserAgent,
		BrowserName:          browserInfo.BrowserName,
		BrowserVersion:       browserInfo.BrowserVersion,
		OSName:               browserInfo.OSName,
		DeviceType:           browserInfo.DeviceType,
		BrowserKey:           browserInfo.Key,
		LoginIP:              input.IP,
		LastIP:               input.IP,
		LastSeenAt:           now,
		CurrentAccessJTI:     accessJTI,
		CurrentAccessExpires: accessExpiresAt,
	}
	refreshRecord := &model.RefreshToken{
		SessionID: sessionID,
		UserID:    userInfo.UserID,
		TokenHash: token.Hash(refreshTokenValue),
		Status:    RefreshStatusActive,
		ExpiresAt: refreshExpiresAt,
	}

	err = s.TxManager.WithinTransaction(ctx, func(tx *gorm.DB) error {
		sessionRepo := s.SessionRepo.WithTx(tx)
		refreshRepo := s.RefreshRepo.WithTx(tx)

		if err := sessionRepo.Create(ctx, session); err != nil {
			return err
		}
		if err := refreshRepo.Create(ctx, refreshRecord); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	_ = s.cacheSession(ctx, session)

	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshTokenValue,
		SessionID:        sessionID,
		AccessExpiresAt:  accessExpiresAt.Unix(),
		RefreshExpiresAt: refreshExpiresAt.Unix(),
	}, nil
}
