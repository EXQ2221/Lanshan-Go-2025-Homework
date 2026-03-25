package convert

import (
	"example.com/micro-auth-demo/auth-service/internal/biz"
	authpb "example.com/micro-auth-demo/auth-service/kitex_gen/auth"
)

func ToTokenPair(pair *biz.TokenPair) *authpb.TokenPair {
	if pair == nil {
		return nil
	}

	return &authpb.TokenPair{
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		SessionId:        pair.SessionID,
		AccessExpiresAt:  pair.AccessExpiresAt,
		RefreshExpiresAt: pair.RefreshExpiresAt,
	}
}

func ToSessionInfos(sessions []biz.SessionView) []*authpb.SessionInfo {
	result := make([]*authpb.SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		result = append(result, &authpb.SessionInfo{
			SessionId:  session.SessionID,
			DeviceId:   session.DeviceID,
			DeviceName: session.DeviceName,
			UserAgent:  session.UserAgent,
			LoginIp:    session.LoginIP,
			LastIp:     session.LastIP,
			Status:     session.Status,
			Current:    session.Current,
			CreatedAt:  session.CreatedAt,
			LastSeenAt: session.LastSeenAt,
		})
	}
	return result
}
