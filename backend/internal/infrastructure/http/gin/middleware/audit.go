package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

func AuditContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		ctx := context.WithValue(c.Request.Context(), "ip", ip)
		ctx = context.WithValue(ctx, "userAgent", userAgent)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
