package biz

import (
	"context"
	"errors"
	"time"

	"example.com/micro-auth-demo/auth-service/internal/pkg/jwt"
	"example.com/micro-auth-demo/auth-service/internal/pkg/token"
	"example.com/micro-auth-demo/auth-service/internal/repository"
	"gorm.io/gorm"
)

func (s *AuthService) ValidateToken(ctx context.Context, accessToken string) (*AuthIdentity, error) {
	claims, err := jwt.Parse(accessToken, s.Secret)
	if err != nil {
		return nil, ErrInvalidAccessToken
	}

	blacklisted, err := s.Cache.IsAccessTokenBlacklisted(ctx, claims.TokenID)
	if err == nil && blacklisted {
		return nil, ErrInvalidAccessToken
	}

	entry, err := s.Cache.GetSession(ctx, claims.SessionID)
	if err != nil {
		entry = nil
	}
	if entry == nil {
		session, err := s.SessionRepo.GetBySessionID(ctx, claims.SessionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrInvalidAccessToken
			}
			return nil, err
		}

		entry = &repository.SessionCacheEntry{
			UserID:           session.UserID,
			Status:           session.Status,
			CurrentAccessJTI: session.CurrentAccessJTI,
		}
		_ = s.Cache.SetSession(ctx, session.SessionID, *entry, s.RefreshTTL)
	}

	if entry.Status != SessionStatusActive {
		return nil, ErrSessionRevoked
	}
	if entry.UserID != claims.UserID || entry.CurrentAccessJTI != claims.TokenID {
		return nil, ErrInvalidAccessToken
	}

	return &AuthIdentity{
		UserID:    claims.UserID,
		SessionID: claims.SessionID,
	}, nil
}

func (s *AuthService) issueAccessToken(userID int64, sessionID string, now time.Time) (string, string, time.Time, error) {
	tokenID, err := token.Generate(16)
	if err != nil {
		return "", "", time.Time{}, err
	}
	expiresAt := now.Add(s.AccessTTL)
	tokenValue, err := jwt.Sign(jwt.NewClaims(userID, sessionID, tokenID, now, expiresAt), s.Secret)
	if err != nil {
		return "", "", time.Time{}, err
	}
	return tokenValue, tokenID, expiresAt, nil
}
