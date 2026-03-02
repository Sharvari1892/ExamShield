package service

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestPasswordHashAndCompare(t *testing.T) {
	auth := NewAuthService("testsecret")

	password := "mypassword"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	if err := auth.CheckPassword(hash, password); err != nil {
		t.Fatalf("password should match")
	}

	if err := auth.CheckPassword(hash, "wrong"); err == nil {
		t.Fatalf("password should not match")
	}
}

func TestAccessTokenGeneration(t *testing.T) {
	auth := NewAuthService("testsecret")

	token, err := auth.GenerateAccessToken("user1", "student")
	if err != nil {
		t.Fatal(err)
	}

	if token == "" {
		t.Fatal("token should not be empty")
	}
}

func TestInvalidSignatureFails(t *testing.T) {
	auth := NewAuthService("correctsecret")

	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "user1",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("wrongsecret"))

	_, err := auth.ValidateToken(token)
	if err == nil {
		t.Fatalf("expected invalid signature to fail")
	}
}
