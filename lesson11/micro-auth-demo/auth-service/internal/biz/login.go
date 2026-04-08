package biz

import (
	"context"
	"time"

	"example.com/micro-auth-demo/auth-service/internal/dal/kafka"
	"example.com/micro-auth-demo/auth-service/internal/dal/model"
	browserprofile "example.com/micro-auth-demo/auth-service/internal/pkg/browser"
	"example.com/micro-auth-demo/auth-service/internal/pkg/geo"
	"example.com/micro-auth-demo/auth-service/internal/pkg/token"
	"example.com/micro-auth-demo/auth-service/internal/repository"
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

	currentGeo, err := s.GeoLocator.Search(input.IP)
	if err != nil {
		// IP 解析失败不影响登录，记日志继续
		currentGeo = &geo.GeoInfo{}
	}

	currentCityKey := currentGeo.GetCityKey() // "重庆|重庆市" 或 "unknown"

	// 3. 从 Redis 取上次登录信息
	lastLoginInfo, _ := s.Cache.GetLastLogin(ctx, userInfo.UserID)

	// 4. 异地检测：有上次记录且不同城市
	if lastLoginInfo != nil && lastLoginInfo.City != "" && lastLoginInfo.City != currentCityKey {
		// 发 Kafka 安全事件
		s.KafkaProducer.SendSecurityEvent(kafka.SecurityEvent{
			EventType:    kafka.EventTypeAbnormalGeoLogin,
			UserID:       userInfo.UserID,
			SessionID:    sessionID,
			IP:           input.IP,
			DeviceID:     input.DeviceID,
			UserAgent:    input.UserAgent,
			Detail:       "abnormal geo location detected",
			OccurAt:      time.Now().Unix(),
			CurrentCity:  currentCityKey,
			PreviousCity: lastLoginInfo.City,
			PreviousIP:   lastLoginInfo.IP,
			PreviousTime: lastLoginInfo.Time,
		})
	}

	// 5. 更新 Redis（无论是否异常都更新为当前）
	s.Cache.SetLastLogin(ctx, userInfo.UserID, &repository.LastLoginInfo{
		City: currentCityKey,
		IP:   input.IP,
		Time: time.Now().Unix(),
	}, 30*24*time.Hour)

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
