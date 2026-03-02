package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/Sharvari1892/examshield/internal/middleware"
	"github.com/Sharvari1892/examshield/internal/service"
)

func setupProtectedRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	authService := service.NewAuthService("testsecret")

	router := gin.New()

	protected := router.Group("/protected")
	protected.Use(middleware.AuthMiddleware(authService))
	protected.GET("", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	return router
}

func TestProtectedNoToken(t *testing.T) {
	router := setupProtectedRouter()

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestProtectedInvalidToken(t *testing.T) {
	router := setupProtectedRouter()

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestProtectedValidToken(t *testing.T) {
	router := setupProtectedRouter()

	auth := service.NewAuthService("testsecret")
	token, _ := auth.GenerateAccessToken("user123", "student")

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
