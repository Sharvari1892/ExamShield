package main

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
	"context"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/Sharvari1892/examshield/internal/middleware"
	"github.com/Sharvari1892/examshield/internal/service"
)

func setupRateLimitRouter(rdb *redis.Client) *gin.Engine {
	gin.SetMode(gin.TestMode)

	auth := service.NewAuthService("testsecret")
	router := gin.New()

	protected := router.Group("/protected")
	protected.Use(
		middleware.AuthMiddleware(auth),
		middleware.RateLimiter(rdb, 5, time.Minute),
	)

	protected.GET("", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	return router
}

func TestRateLimiterSequential(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	rdb.FlushAll(context.Background())

	router := setupRateLimitRouter(rdb)

	auth := service.NewAuthService("testsecret")
	token, _ := auth.GenerateAccessToken("user1", "student")

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	}

	// 6th request should fail
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", w.Code)
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	rdb.FlushAll(context.Background())

	router := setupRateLimitRouter(rdb)

	auth := service.NewAuthService("testsecret")
	token, _ := auth.GenerateAccessToken("user2", "student")

	var wg sync.WaitGroup
	successCount := 0
	mu := sync.Mutex{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if successCount != 5 {
		t.Fatalf("expected 5 successful requests, got %d", successCount)
	}
}
