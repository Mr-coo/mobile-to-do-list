package config

import (
	"os"
	"strings"
)

type Config struct {
	Port        string
	MySQLDSN    string
	RedisAddr   string
	CORSOrigins []string
}

func Load() Config {
	return Config{
		Port:        getEnv("PORT", "8080"),
		MySQLDSN:    os.Getenv("MYSQL_DSN"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		CORSOrigins: strings.Split(getEnv("CORS_ORIGINS", "http://localhost:8081"), ","),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
