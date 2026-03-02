package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBUrl     string
	RedisAddr string
	JWTSecret string
	Port      string
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		DBUrl:     getEnv("DB_URL", "postgres://exam:exam@localhost:5432/exam?sslmode=disable"),
		RedisAddr: getEnv("REDIS_ADDR", "localhost:6379"),
		JWTSecret: getEnv("JWT_SECRET", "supersecretkey"),
		Port:      getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
