package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Sharvari1892/examshield/internal/repository"
	"github.com/Sharvari1892/examshield/internal/service"
)

func setupRouter(t *testing.T) *gin.Engine {
	gin.SetMode(gin.TestMode)

	db, err := pgxpool.New(context.Background(),
		"postgres://exam:exam@localhost:5432/exam?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}

	authService := service.NewAuthService("testsecret")
	userRepo := repository.NewUserRepository(db)

	router := gin.Default()

	router.POST("/login", func(c *gin.Context) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "invalid request"})
			return
		}

		user, err := userRepo.GetByEmail(c, req.Email)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}

		if err := authService.CheckPassword(user.PasswordHash, req.Password); err != nil {
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}

		accessToken, _ := authService.GenerateAccessToken(user.ID, user.Role)
		refreshToken, _ := authService.GenerateRefreshToken(user.ID)

		c.JSON(200, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})

	return router
}

func TestLoginSuccess(t *testing.T) {
	router := setupRouter(t)

	db, _ := pgxpool.New(context.Background(),
		"postgres://exam:exam@localhost:5432/exam?sslmode=disable")

	auth := service.NewAuthService("testsecret")
	hash, _ := auth.HashPassword("mypassword")

	userRepo := repository.NewUserRepository(db)
	userRepo.CreateUser(context.Background(), "test@example.com", hash, "student")

	body := map[string]string{
		"email":    "test@example.com",
		"password": "mypassword",
	}

	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	router := setupRouter(t)

	body := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpass",
	}

	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestLoginUserNotFound(t *testing.T) {
	router := setupRouter(t)

	body := map[string]string{
		"email":    "nouser@example.com",
		"password": "pass",
	}

	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

