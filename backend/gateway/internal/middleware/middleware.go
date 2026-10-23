package middleware

import (
	"net/http"
	"strings"

	"aziz.dev/gateway/internal/config"
	"aziz.dev/gateway/internal/security"
	"github.com/gin-gonic/gin"
)

const UserIDKey = "userID"

func AccessTokenMiddleware(cfg *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("Authorization"))

		// Accept both "Bearer <token>" and a raw token.
		if len(token) > 7 && strings.EqualFold(token[:7], "Bearer ") {
			token = strings.TrimSpace(token[7:])
		}
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		claims, err := security.ValidateToken(cfg, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(UserIDKey, claims.UserID.String())
		c.Next()
	}
}