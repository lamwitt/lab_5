package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	Port       string

	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessExpiry  time.Duration
	JWTRefreshExpiry time.Duration

	OAuthYandexClientID     string
	OAuthYandexClientSecret string
	OAuthYandexCallbackURL  string

	SwaggerEnabled bool

	RedisHost     string
	RedisPort     string
	RedisPassword string
	CacheTTL      int
}

func Load() *Config {
	return &Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "student"),
		DBPassword: getEnv("DB_PASSWORD", "student_secure_password"),
		DBName:     getEnv("DB_NAME", "wp_labs"),
		Port:       getEnv("PORT", "4200"),

		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "change_me_access_secret"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "change_me_refresh_secret"),
		JWTAccessExpiry:  parseDuration(getEnv("JWT_ACCESS_EXPIRATION", "15m")),
		JWTRefreshExpiry: parseDuration(getEnv("JWT_REFRESH_EXPIRATION", "7d")),

		OAuthYandexClientID:     getEnv("CLIENT_ID", ""),
		OAuthYandexClientSecret: getEnv("CLIENT_SECRET", ""),
		OAuthYandexCallbackURL:  getEnv("CALLBACK_URL", "http://localhost:4200/auth/oauth/yandex/callback"),

		SwaggerEnabled: getEnv("SWAGGER_ENABLED", "false") == "true",

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		CacheTTL:      getEnvInt("CACHE_TTL_DEFAULT", 300),
	}
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName,
	)
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return defaultValue
}

func parseDuration(s string) time.Duration {
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err == nil {
			return time.Duration(days) * 24 * time.Hour
		}
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 15 * time.Minute
	}
	return d
}
