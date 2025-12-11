package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	ICE      ICEConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret           string
	Expiry           time.Duration
	RefreshExpiry    time.Duration
	PasswordResetExp time.Duration
}

type ICEConfig struct {
	STUNServers    []string
	TURNServers    []string
	TURNUsername   string
	TURNCredential string
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode,
	)
}

func Load() *Config {
	// Load .env file if exists (ignores error if not found)
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "3000"),
			Mode: getEnv("GIN_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "streamz"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:           getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
			Expiry:           getEnvDuration("JWT_EXPIRY", 15*time.Minute),
			RefreshExpiry:    getEnvDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
			PasswordResetExp: getEnvDuration("PASSWORD_RESET_EXPIRY", 1*time.Hour),
		},
		ICE: ICEConfig{
			STUNServers: getEnvSlice("ICE_STUN_SERVERS", []string{
				"stun:stun.l.google.com:19302",
				"stun:openrelay.metered.ca:80",
			}),
			TURNServers: getEnvSlice("ICE_TURN_SERVERS", []string{
				"turn:openrelay.metered.ca:80",
				"turn:openrelay.metered.ca:443",
				"turn:openrelay.metered.ca:443?transport=tcp",
			}),
			TURNUsername:   getEnv("ICE_TURN_USERNAME", "openrelayproject"),
			TURNCredential: getEnv("ICE_TURN_CREDENTIAL", "openrelayproject"),
		},
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if val := os.Getenv(key); val != "" {
		return strings.Split(val, ",")
	}
	return fallback
}
