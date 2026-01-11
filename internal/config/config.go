package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Email    EmailConfig
}

type ServerConfig struct {
	Host          string
	Port          int
	Secure        bool   // Use HTTPS-only cookies
	Environment   string // "development", "production", "test"
	Debug         bool
	DebugMaxChars int
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type EmailConfig struct {
	Provider     string // "resend", "smtp", "console"
	FromAddress  string
	FromName     string
	BaseURL      string // Application base URL for links
	ResendAPIKey string
	// SMTP settings (for Mailpit in local dev)
	SMTPHost string
	SMTPPort int
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host:          getEnv("SERVER_HOST", "0.0.0.0"),
			Port:          getEnvInt("SERVER_PORT", 8080),
			Secure:        getEnvBool("SERVER_SECURE", false),
			Environment:   getEnv("APP_ENV", "development"),
			Debug:         getEnvBool("DEBUG", false),
			DebugMaxChars: getEnvInt("DEBUG_LOG_MAX_CHARS", 8000),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "__TEMPLATE_PROJECT_SLUG__"),
			Password: getEnv("DB_PASSWORD", "__TEMPLATE_PROJECT_SLUG__"),
			DBName:   getEnv("DB_NAME", "__TEMPLATE_PROJECT_SLUG__"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Email: EmailConfig{
			Provider:     getEnv("EMAIL_PROVIDER", "console"),
			FromAddress:  getEnv("EMAIL_FROM_ADDRESS", "__TEMPLATE_EMAIL_FROM_ADDRESS__"),
			FromName:     getEnv("EMAIL_FROM_NAME", "__TEMPLATE_EMAIL_FROM_NAME__"),
			BaseURL:      getEnv("APP_BASE_URL", "__TEMPLATE_APP_BASE_URL__"),
			ResendAPIKey: getEnv("RESEND_API_KEY", ""),
			SMTPHost:     getEnv("SMTP_HOST", "localhost"),
			SMTPPort:     getEnvInt("SMTP_PORT", 1025),
		},
}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvNonEmpty(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		if strings.TrimSpace(value) != "" {
			return value
		}
		return defaultValue
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
