package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// MaxBodyBytes limits request bodies for non-multipart endpoints. Multipart
// uploads keep their dedicated size limit in the upload handler.
func MaxBodyBytes(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil && !strings.HasPrefix(strings.ToLower(c.GetHeader("Content-Type")), "multipart/form-data") {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	}
}
