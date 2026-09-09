package media

import "net/http"

// DetectContentType provides a shared media type detector for future processors.
func DetectContentType(data []byte) string {
	return http.DetectContentType(data)
}
