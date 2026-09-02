package middleware

import (
	"errors"
	"net/http"
	"strings"

	"notell/services"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserIDKey = "userId"
	ContextEmailKey  = "email"
)

func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header != "" {
			parts := strings.Fields(header)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid authorization header format"})
				return
			}
			tokenString = parts[1]
		}

		if tokenString == "" {
			if cookie, err := c.Cookie("auth_token"); err == nil {
				tokenString = strings.TrimSpace(cookie)
			}
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Authentication required"})
			return
		}

		claims, err := services.ValidateToken(tokenString, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid or expired token"})
			return
		}

		if claims.UserID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid token claims"})
			return
		}

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextEmailKey, claims.Email)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uint, error) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, errors.New("user does not exist in context")
	}

	userID, ok := val.(uint)
	if !ok || userID == 0 {
		return 0, errors.New("invalid user ID type in context")
	}

	return userID, nil
}
