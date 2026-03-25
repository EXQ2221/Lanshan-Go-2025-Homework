package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"strings"

	"example.com/micro-auth-demo/gateway/internal/middleware"
	"example.com/micro-auth-demo/gateway/internal/model"
	"example.com/micro-auth-demo/gateway/internal/rpc"
	authpb "example.com/micro-auth-demo/gateway/kitex_gen/auth"
	"github.com/gin-gonic/gin"
)

func Login(ctx *gin.Context) {
	var req model.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, 400, "invalid request body")
		return
	}

	client, err := rpc.AuthClient()
	if err != nil {
		writeError(ctx, 500, err.Error())
		return
	}

	if req.DeviceID == "" {
		req.DeviceID = newID()
	}
	if req.DeviceName == "" {
		req.DeviceName = "web-client"
	}

	resp, err := client.Login(ctx.Request.Context(), &authpb.LoginRequest{
		Email:      req.Email,
		Password:   req.Password,
		DeviceId:   req.DeviceID,
		DeviceName: req.DeviceName,
		UserAgent:  ctx.Request.UserAgent(),
		Ip:         clientIP(ctx),
	})
	if err != nil {
		writeError(ctx, statusFromMessage(err.Error()), presentableMessage(err.Error()))
		return
	}

	writeJSON(ctx, 200, model.APIResponse{
		Code:    0,
		Message: "success",
		Data: model.TokenPair{
			AccessToken:      resp.AccessToken,
			RefreshToken:     resp.RefreshToken,
			SessionID:        resp.SessionId,
			AccessExpiresAt:  resp.AccessExpiresAt,
			RefreshExpiresAt: resp.RefreshExpiresAt,
		},
	})
}

func Refresh(ctx *gin.Context) {
	var req model.RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, 400, "invalid request body")
		return
	}

	client, err := rpc.AuthClient()
	if err != nil {
		writeError(ctx, 500, err.Error())
		return
	}

	resp, err := client.RefreshToken(ctx.Request.Context(), &authpb.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
		DeviceId:     req.DeviceID,
		UserAgent:    ctx.Request.UserAgent(),
		Ip:           clientIP(ctx),
	})
	if err != nil {
		writeError(ctx, statusFromMessage(err.Error()), presentableMessage(err.Error()))
		return
	}

	writeJSON(ctx, 200, model.APIResponse{
		Code:    0,
		Message: "success",
		Data: model.TokenPair{
			AccessToken:      resp.AccessToken,
			RefreshToken:     resp.RefreshToken,
			SessionID:        resp.SessionId,
			AccessExpiresAt:  resp.AccessExpiresAt,
			RefreshExpiresAt: resp.RefreshExpiresAt,
		},
	})
}

func Logout(ctx *gin.Context) {
	authCtx, ok := middleware.GetAuthContext(ctx)
	if !ok {
		writeError(ctx, 401, "missing auth context")
		return
	}

	client, err := rpc.AuthClient()
	if err != nil {
		writeError(ctx, 500, err.Error())
		return
	}

	resp, err := client.Logout(ctx.Request.Context(), &authpb.LogoutRequest{
		AccessToken: authCtx.AccessToken,
	})
	if err != nil {
		writeError(ctx, statusFromMessage(err.Error()), presentableMessage(err.Error()))
		return
	}
	if !resp.Success {
		writeError(ctx, statusFromMessage(resp.Message), presentableMessage(resp.Message))
		return
	}

	writeJSON(ctx, 200, model.APIResponse{Code: 0, Message: "success"})
}

func LogoutAll(ctx *gin.Context) {
	authCtx, ok := middleware.GetAuthContext(ctx)
	if !ok {
		writeError(ctx, 401, "missing auth context")
		return
	}

	var req model.PasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, 400, "invalid request body")
		return
	}

	client, err := rpc.AuthClient()
	if err != nil {
		writeError(ctx, 500, err.Error())
		return
	}

	resp, err := client.LogoutAll(ctx.Request.Context(), &authpb.LogoutAllRequest{
		UserId:   authCtx.UserID,
		Password: req.Password,
	})
	if err != nil {
		writeError(ctx, statusFromMessage(err.Error()), presentableMessage(err.Error()))
		return
	}
	if !resp.Success {
		writeError(ctx, statusFromMessage(resp.Message), presentableMessage(resp.Message))
		return
	}

	writeJSON(ctx, 200, model.APIResponse{Code: 0, Message: "success"})
}

func ListSessions(ctx *gin.Context) {
	authCtx, ok := middleware.GetAuthContext(ctx)
	if !ok {
		writeError(ctx, 401, "missing auth context")
		return
	}

	client, err := rpc.AuthClient()
	if err != nil {
		writeError(ctx, 500, err.Error())
		return
	}

	resp, err := client.ListSessions(ctx.Request.Context(), &authpb.ListSessionsRequest{
		UserId:           authCtx.UserID,
		CurrentSessionId: authCtx.SessionID,
	})
	if err != nil {
		writeError(ctx, statusFromMessage(err.Error()), presentableMessage(err.Error()))
		return
	}

	sessions := make([]model.SessionInfo, 0, len(resp.Sessions))
	for _, session := range resp.Sessions {
		sessions = append(sessions, model.SessionInfo{
			SessionID:  session.SessionId,
			DeviceID:   session.DeviceId,
			DeviceName: session.DeviceName,
			UserAgent:  session.UserAgent,
			LoginIP:    session.LoginIp,
			LastIP:     session.LastIp,
			Status:     session.Status,
			Current:    session.Current,
			CreatedAt:  session.CreatedAt,
			LastSeenAt: session.LastSeenAt,
		})
	}

	writeJSON(ctx, 200, model.APIResponse{
		Code:    0,
		Message: "success",
		Data:    sessions,
	})
}

func RevokeSession(ctx *gin.Context) {
	authCtx, ok := middleware.GetAuthContext(ctx)
	if !ok {
		writeError(ctx, 401, "missing auth context")
		return
	}

	var req model.RevokeSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		writeError(ctx, 400, "invalid request body")
		return
	}

	client, err := rpc.AuthClient()
	if err != nil {
		writeError(ctx, 500, err.Error())
		return
	}

	resp, err := client.RevokeSession(ctx.Request.Context(), &authpb.RevokeSessionRequest{
		UserId:    authCtx.UserID,
		SessionId: req.SessionID,
		Password:  req.Password,
	})
	if err != nil {
		writeError(ctx, statusFromMessage(err.Error()), presentableMessage(err.Error()))
		return
	}
	if !resp.Success {
		writeError(ctx, statusFromMessage(resp.Message), presentableMessage(resp.Message))
		return
	}

	writeJSON(ctx, 200, model.APIResponse{Code: 0, Message: "success"})
}

func statusFromMessage(message string) int {
	switch {
	case strings.Contains(message, "unauthorized:"):
		return 401
	case strings.Contains(message, "forbidden:"):
		return 403
	case strings.Contains(message, "not_found:"):
		return 404
	default:
		return 500
	}
}

func presentableMessage(message string) string {
	for _, prefix := range []string{"unauthorized: ", "forbidden: ", "not_found: "} {
		if idx := strings.Index(message, prefix); idx >= 0 {
			return message[idx+len(prefix):]
		}
	}
	return message
}

func clientIP(ctx *gin.Context) string {
	if forwarded := ctx.GetHeader("X-Forwarded-For"); forwarded != "" {
		parts := strings.Split(forwarded, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	host, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
	if err == nil {
		return host
	}
	return ctx.Request.RemoteAddr
}

func newID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "web-client"
	}
	return hex.EncodeToString(buf)
}
