package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"notell/services"

	"github.com/gin-gonic/gin"
)

func runAuthRequest(t *testing.T, method, path, authorization, cookie string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET(path, AuthRequired("test-secret"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"userId": c.MustGet(ContextUserIDKey)})
	})

	req := httptest.NewRequest(method, path, nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: "auth_token", Value: cookie})
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestAuthRequiredRejectsMissingCredentials(t *testing.T) {
	w := runAuthRequest(t, http.MethodGet, "/protected", "", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthRequiredRejectsMalformedBearerHeader(t *testing.T) {
	cases := []string{
		"Basic abc",
		"Bearer",
		"Bearer    ",
		"Bearer token extra",
	}

	for _, header := range cases {
		t.Run(header, func(t *testing.T) {
			w := runAuthRequest(t, http.MethodGet, "/protected", header, "")
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestAuthRequiredAcceptsValidBearerToken(t *testing.T) {
	token, err := services.GenerateToken(42, "user@example.com", "test-secret")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	w := runAuthRequest(t, http.MethodGet, "/protected", "Bearer "+token, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthRequiredAcceptsValidCookie(t *testing.T) {
	token, err := services.GenerateToken(7, "cookie@example.com", "test-secret")
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	w := runAuthRequest(t, http.MethodGet, "/protected", "", token)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuthRequiredRejectsInvalidToken(t *testing.T) {
	w := runAuthRequest(t, http.MethodGet, "/protected", "Bearer invalid-token", "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestGetUserIDValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	if _, err := GetUserID(c); err == nil {
		t.Fatal("GetUserID() accepted missing context")
	}

	c.Set(ContextUserIDKey, uint(0))
	if _, err := GetUserID(c); err == nil {
		t.Fatal("GetUserID() accepted zero user ID")
	}

	c.Set(ContextUserIDKey, "42")
	if _, err := GetUserID(c); err == nil {
		t.Fatal("GetUserID() accepted wrong user ID type")
	}

	c.Set(ContextUserIDKey, uint(42))
	id, err := GetUserID(c)
	if err != nil {
		t.Fatalf("GetUserID() error = %v", err)
	}
	if id != 42 {
		t.Fatalf("GetUserID() = %d, want 42", id)
	}
}
