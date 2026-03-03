package middleware

import (
	"net/http"

	"github.com/diogenes-moreira/creditos/backend/internal/domain/model"
	"github.com/diogenes-moreira/creditos/backend/internal/domain/port"
	"github.com/gin-gonic/gin"
)

func RequireRole(userRepo port.UserRepository, roles ...model.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		firebaseUID, exists := c.Get("firebaseUID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		user, err := userRepo.FindByFirebaseUID(c.Request.Context(), firebaseUID.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		if !user.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "account is deactivated"})
			return
		}

		allowed := false
		for _, role := range roles {
			if user.Role == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Set("userID", user.ID)
		c.Set("userRole", user.Role)
		c.Next()
	}
}
