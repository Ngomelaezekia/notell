package services

import (
	"path/filepath"
	"strings"
)

func MediaObjectKey(filename string) string {
	return "uploads/" + filepath.Base(filename)
}

func MediaPublicURL(baseURL, key string) string {
	return strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(key, "/")
}
