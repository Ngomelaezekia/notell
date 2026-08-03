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

// AuthRequired verifies JWT tokens from either:
// 1. Authorization: Bearer <token>
// 2. HttpOnly cookie: auth_token
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {

		var tokenString string

		// -------------------------------------------------
		// Try Authorization Header
		// -------------------------------------------------

		header := c.GetHeader("Authorization")

		if header != "" {
			parts := strings.SplitN(header, " ", 2)

			if len(parts) != 2 || parts[0] != "Bearer" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "Invalid authorization header format",
				})
				return
			}

			tokenString = parts[1]
		}

		// -------------------------------------------------
		// If no Authorization header, try HttpOnly Cookie
		// -------------------------------------------------

		if tokenString == "" {
			if cookie, err := c.Cookie("auth_token"); err == nil {
				tokenString = cookie
			}
		}

		// -------------------------------------------------
		// No authentication found
		// -------------------------------------------------

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Authentication required",
			})
			return
		}

		// -------------------------------------------------
		// Validate JWT
		// -------------------------------------------------

		claims, err := services.ValidateToken(tokenString, jwtSecret)

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid or expired token",
			})
			return
		}

		// -------------------------------------------------
		// Save User Context
		// -------------------------------------------------

		c.Set(ContextUserIDKey, claims.UserID)
		c.Set(ContextEmailKey, claims.Email)

		c.Next()
	}
}

// GetUserID safely extracts the authenticated user ID.
func GetUserID(c *gin.Context) (uint, error) {
	val, exists := c.Get(ContextUserIDKey)
	if !exists {
		return 0, errors.New("user does not exist in context")
	}

	userID, ok := val.(uint)
	if !ok {
		return 0, errors.New("invalid user ID type in context")
	}

	return userID, nil
}
