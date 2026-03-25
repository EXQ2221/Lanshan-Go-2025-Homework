package middleware

import (
	"strings"

	"example.com/micro-auth-demo/gateway/internal/rpc"
	authpb "example.com/micro-auth-demo/gateway/kitex_gen/auth"
	"github.com/gin-gonic/gin"
)

const authContextKey = "auth-context"

type AuthContext struct {
	UserID      int64
	SessionID   string
	AccessToken string
}

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token := bearerToken(ctx.GetHeader("Authorization"))
		if token == "" {
			ctx.AbortWithStatusJSON(401, gin.H{
				"code":    401,
				"message": "missing authorization header",
			})
			return
		}

		client, err := rpc.AuthClient()
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}

		resp, err := client.ValidateToken(ctx.Request.Context(), &authpb.ValidateTokenRequest{
			AccessToken: token,
		})
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{
				"code":    500,
				"message": err.Error(),
			})
			return
		}
		if !resp.Valid {
			ctx.AbortWithStatusJSON(401, gin.H{
				"code":    401,
				"message": resp.Reason,
			})
			return
		}

		ctx.Set(authContextKey, AuthContext{
			UserID:      resp.UserId,
			SessionID:   resp.SessionId,
			AccessToken: token,
		})
		ctx.Next()
	}
}

func GetAuthContext(ctx *gin.Context) (AuthContext, bool) {
	value, ok := ctx.Get(authContextKey)
	if !ok {
		return AuthContext{}, false
	}

	authCtx, ok := value.(AuthContext)
	if !ok {
		return AuthContext{}, false
	}

	return authCtx, true
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}

	return strings.TrimSpace(strings.TrimPrefix(header, "Bearer"))
}
