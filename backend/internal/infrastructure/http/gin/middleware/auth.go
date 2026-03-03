package middleware

import (
	"net/http"
	"strings"

	"github.com/diogenes-moreira/creditos/backend/internal/infrastructure/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authService *auth.LocalAuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		user, err := authService.VerifyToken(c.Request.Context(), parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("firebaseUID", user.UID)
		c.Set("email", user.Email)
		c.Next()
	}
}
