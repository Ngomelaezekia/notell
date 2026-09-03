package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// CSRFProtection protects cookie-authenticated state-changing requests by
// requiring the browser Origin to match the configured frontend origin.
// Bearer-token clients are not subject to this check because they do not
// rely on ambient browser credentials.
func CSRFProtection(frontendURL string) gin.HandlerFunc {
	allowedOrigin := originFromURL(frontendURL)

	return func(c *gin.Context) {
		if isSafeMethod(c.Request.Method) || strings.TrimSpace(c.GetHeader("Authorization")) != "" {
			c.Next()
			return
		}

		if _, err := c.Cookie("auth_token"); err != nil {
			c.Next()
			return
		}

		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin == "" || !strings.EqualFold(origin, allowedOrigin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "invalid request origin"})
			return
		}

		c.Next()
	}
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func originFromURL(value string) string {
	parsed, err := url.Parse(strings.TrimSpace(value))
	if err != nil {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}
