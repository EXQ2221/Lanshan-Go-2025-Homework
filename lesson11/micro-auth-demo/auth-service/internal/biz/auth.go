package biz

import (
	"context"
	"errors"
	"time"

	"example.com/micro-auth-demo/auth-service/internal/dal/model"
	"example.com/micro-auth-demo/auth-service/internal/pkg/jwt"
	"example.com/micro-auth-demo/auth-service/internal/pkg/token"
	"example.com/micro-auth-demo/auth-service/internal/repository"
	"example.com/micro-auth-demo/auth-service/internal/rpc"
	"gorm.io/gorm"
)

type TokenPair struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	SessionID        string `json:"session_id"`
	AccessExpiresAt  int64  `json:"access_expires_at"`
	RefreshExpiresAt int64  `json:"refresh_expires_at"`
}

type AuthService struct {
	SessionRepo repository.SessionRepository
	RefreshRepo repository.RefreshTokenRepository
	EventRepo   repository.SecurityEventRepository
	TxManager   repository.TxManager
	Cache       repository.AuthCache
	UserClient  rpc.UserClient
	Secret      string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
}

func NewAuthService(
	sessionRepo repository.SessionRepository,
	refreshRepo repository.RefreshTokenRepository,
	eventRepo repository.SecurityEventRepository,
	txManager repository.TxManager,
	cache repository.AuthCache,
	userClient rpc.UserClient,
	secret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		SessionRepo: sessionRepo,
		RefreshRepo: refreshRepo,
		EventRepo:   eventRepo,
		TxManager:   txManager,
		Cache:       cache,
		UserClient:  userClient,
		Secret:      secret,
		AccessTTL:   accessTTL,
		RefreshTTL:  refreshTTL,
	}
}

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
	session := &model.Session{
		SessionID:            sessionID,
		UserID:               userInfo.UserID,
		Status:               SessionStatusActive,
		DeviceID:             coalesce(input.DeviceID, sessionID),
		DeviceName:           coalesce(input.DeviceName, "unknown-device"),
		UserAgent:            input.UserAgent,
		LoginIP:              input.IP,
		LastIP:               input.IP,
		LastSeenAt:           now,
		CurrentAccessJTI:     accessJTI,
		CurrentAccessExpires: accessExpiresAt,
	}
	if err := s.SessionRepo.Create(ctx, session); err != nil {
		return nil, err
	}

	if err := s.RefreshRepo.Create(ctx, &model.RefreshToken{
		SessionID: sessionID,
		UserID:    userInfo.UserID,
		TokenHash: token.Hash(refreshTokenValue),
		Status:    RefreshStatusActive,
		ExpiresAt: refreshExpiresAt,
	}); err != nil {
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

func (s *AuthService) Refresh(ctx context.Context, input RefreshInput) (*TokenPair, error) {
	now := time.Now()
	refreshHash := token.Hash(input.RefreshToken)

	var (
		pair         *TokenPair
		cacheSession *model.Session
		reuseRecord  *model.RefreshToken
		finalErr     error
	)

	err := s.TxManager.WithinTransaction(ctx, func(tx *gorm.DB) error {
		refreshRepo := s.RefreshRepo.WithTx(tx)
		sessionRepo := s.SessionRepo.WithTx(tx)
		eventRepo := s.EventRepo.WithTx(tx)

		record, err := refreshRepo.GetByTokenHashForUpdate(ctx, refreshHash)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				finalErr = ErrInvalidRefreshToken
				return nil
			}
			return err
		}

		session, err := sessionRepo.GetBySessionIDForUpdate(ctx, record.SessionID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				finalErr = ErrSessionNotFound
				return nil
			}
			return err
		}

		if record.Status != RefreshStatusActive {
			reuseRecord = record
			finalErr = ErrRefreshReuse
			return nil
		}
		if now.After(record.ExpiresAt) {
			record.Status = RefreshStatusRevoked
			record.RevokedAt = &now
			record.RevokeReason = "expired"
			if err := refreshRepo.Update(ctx, record); err != nil {
				return err
			}
			finalErr = ErrInvalidRefreshToken
			return nil
		}
		if session.Status != SessionStatusActive {
			finalErr = ErrSessionRevoked
			return nil
		}
		if session.DeviceID != "" && input.DeviceID != "" && session.DeviceID != input.DeviceID {
			if err := s.recordEventWithRepo(ctx, eventRepo, session.UserID, session.SessionID, "device_mismatch", input.IP, input.DeviceID, input.UserAgent, "refresh attempted from another device id"); err != nil {
				return err
			}
			if err := s.revokeSessionWithRepos(ctx, session, "device_mismatch", sessionRepo, refreshRepo); err != nil {
				return err
			}
			cacheSession = session
			finalErr = ErrDeviceMismatch
			return nil
		}
		if session.UserAgent != "" && input.UserAgent != "" && session.UserAgent != input.UserAgent {
			if err := s.recordEventWithRepo(ctx, eventRepo, session.UserID, session.SessionID, "browser_mismatch", input.IP, input.DeviceID, input.UserAgent, "refresh attempted from another browser"); err != nil {
				return err
			}
		}
		if session.LastIP != "" && input.IP != "" && session.LastIP != input.IP {
			if err := s.recordEventWithRepo(ctx, eventRepo, session.UserID, session.SessionID, "ip_changed", input.IP, input.DeviceID, input.UserAgent, "refresh token used from a new ip"); err != nil {
				return err
			}
		}

		newRefreshToken, err := NewRefreshToken()
		if err != nil {
			return err
		}
		accessToken, accessJTI, accessExpiresAt, err := s.issueAccessToken(session.UserID, session.SessionID, now)
		if err != nil {
			return err
		}

		record.Status = RefreshStatusUsed
		record.UsedAt = &now
		record.LastUsedIP = input.IP
		record.LastUsedUserAgent = input.UserAgent
		record.RotatedTo = token.Hash(newRefreshToken)
		if err := refreshRepo.Update(ctx, record); err != nil {
			return err
		}

		refreshExpiresAt := now.Add(s.RefreshTTL)
		if err := refreshRepo.Create(ctx, &model.RefreshToken{
			SessionID: session.SessionID,
			UserID:    session.UserID,
			TokenHash: token.Hash(newRefreshToken),
			Status:    RefreshStatusActive,
			ExpiresAt: refreshExpiresAt,
		}); err != nil {
			return err
		}

		session.LastSeenAt = now
		session.LastIP = input.IP
		session.CurrentAccessJTI = accessJTI
		session.CurrentAccessExpires = accessExpiresAt
		if err := sessionRepo.Update(ctx, session); err != nil {
			return err
		}

		cacheSession = session
		pair = &TokenPair{
			AccessToken:      accessToken,
			RefreshToken:     newRefreshToken,
			SessionID:        session.SessionID,
			AccessExpiresAt:  accessExpiresAt.Unix(),
			RefreshExpiresAt: refreshExpiresAt.Unix(),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if reuseRecord != nil {
		_ = s.handleRefreshReuse(ctx, reuseRecord, input)
	}
	if cacheSession != nil {
		if cacheSession.Status == SessionStatusRevoked {
			s.afterSessionRevoked(ctx, cacheSession)
		} else {
			_ = s.cacheSession(ctx, cacheSession)
		}
	}
	if finalErr != nil {
		return nil, finalErr
	}

	return pair, nil
}

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

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	claims, err := jwt.Parse(accessToken, s.Secret)
	if err != nil {
		return ErrInvalidAccessToken
	}

	session, err := s.SessionRepo.GetBySessionID(ctx, claims.SessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return err
	}

	return s.revokeSession(ctx, session, "logout")
}

func (s *AuthService) LogoutAll(ctx context.Context, userID int64, password string) error {
	ok, err := s.UserClient.CheckPassword(ctx, userID, password)
	if err != nil {
		return err
	}
	if !ok {
		return ErrPasswordConfirm
	}
	return s.revokeAllSessions(ctx, userID, "logout_all")
}

func (s *AuthService) ListSessions(ctx context.Context, userID int64, currentSessionID string) ([]SessionView, error) {
	sessions, err := s.SessionRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]SessionView, 0, len(sessions))
	for _, session := range sessions {
		if session.Status == SessionStatusRevoked {
			continue
		}
		result = append(result, SessionView{
			SessionID:  session.SessionID,
			DeviceID:   session.DeviceID,
			DeviceName: session.DeviceName,
			UserAgent:  session.UserAgent,
			LoginIP:    session.LoginIP,
			LastIP:     session.LastIP,
			Status:     session.Status,
			Current:    session.SessionID == currentSessionID,
			CreatedAt:  session.CreatedAt.Unix(),
			LastSeenAt: session.LastSeenAt.Unix(),
		})
	}

	return result, nil
}

func (s *AuthService) RevokeSession(ctx context.Context, userID int64, sessionID, password string) error {
	ok, err := s.UserClient.CheckPassword(ctx, userID, password)
	if err != nil {
		return err
	}
	if !ok {
		return ErrPasswordConfirm
	}

	session, err := s.SessionRepo.GetBySessionID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSessionNotFound
		}
		return err
	}
	if session.UserID != userID {
		return ErrSessionNotFound
	}

	return s.revokeSession(ctx, session, "revoke_session")
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

func (s *AuthService) revokeAllSessions(ctx context.Context, userID int64, reason string) error {
	sessions, err := s.SessionRepo.ListByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for i := range sessions {
		if sessions[i].Status == SessionStatusRevoked {
			continue
		}
		if err := s.revokeSession(ctx, &sessions[i], reason); err != nil {
			return err
		}
	}

	return nil
}

func (s *AuthService) revokeSession(ctx context.Context, session *model.Session, reason string) error {
	err := s.TxManager.WithinTransaction(ctx, func(tx *gorm.DB) error {
		sessionRepo := s.SessionRepo.WithTx(tx)
		refreshRepo := s.RefreshRepo.WithTx(tx)
		return s.revokeSessionWithRepos(ctx, session, reason, sessionRepo, refreshRepo)
	})
	if err != nil {
		return err
	}

	s.afterSessionRevoked(ctx, session)
	return nil
}

func (s *AuthService) handleRefreshReuse(ctx context.Context, record *model.RefreshToken, input RefreshInput) error {
	_ = s.recordEvent(ctx, record.UserID, record.SessionID, "refresh_token_reuse", input.IP, input.DeviceID, input.UserAgent, "used or revoked refresh token was presented again")
	return s.revokeAllSessions(ctx, record.UserID, "refresh_token_reuse")
}

func (s *AuthService) cacheSession(ctx context.Context, session *model.Session) error {
	return s.Cache.SetSession(ctx, session.SessionID, repository.SessionCacheEntry{
		UserID:           session.UserID,
		Status:           session.Status,
		CurrentAccessJTI: session.CurrentAccessJTI,
	}, s.RefreshTTL)
}

func (s *AuthService) recordEvent(ctx context.Context, userID int64, sessionID, eventType, ip, deviceID, userAgent, detail string) error {
	return s.recordEventWithRepo(ctx, s.EventRepo, userID, sessionID, eventType, ip, deviceID, userAgent, detail)
}

func (s *AuthService) recordEventWithRepo(ctx context.Context, eventRepo repository.SecurityEventRepository, userID int64, sessionID, eventType, ip, deviceID, userAgent, detail string) error {
	return eventRepo.Create(ctx, &model.SecurityEvent{
		UserID:    userID,
		SessionID: sessionID,
		EventType: eventType,
		IP:        ip,
		DeviceID:  deviceID,
		UserAgent: userAgent,
		Detail:    detail,
		CreatedAt: time.Now(),
	})
}

func (s *AuthService) revokeSessionWithRepos(ctx context.Context, session *model.Session, reason string, sessionRepo repository.SessionRepository, refreshRepo repository.RefreshTokenRepository) error {
	now := time.Now()
	session.Status = SessionStatusRevoked
	session.RevokedAt = &now
	session.RevokeReason = reason

	if err := sessionRepo.Update(ctx, session); err != nil {
		return err
	}
	return refreshRepo.RevokeActiveBySessionID(ctx, session.SessionID, reason, now)
}

func (s *AuthService) afterSessionRevoked(ctx context.Context, session *model.Session) {
	if ttl := time.Until(session.CurrentAccessExpires); ttl > 0 {
		_ = s.Cache.BlacklistAccessToken(ctx, session.CurrentAccessJTI, ttl)
	}
	_ = s.cacheSession(ctx, session)
}

func coalesce(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
