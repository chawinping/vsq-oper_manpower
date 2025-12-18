package config

import (
	"os"
	"strings"
)

type Config struct {
	Database      DatabaseConfig
	Port          string
	SessionSecret string
	CORS          CORSConfig
	MCP           MCPConfig
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type CORSConfig struct {
	AllowedOrigins []string
}

type MCPConfig struct {
	ServerURL string
	APIKey    string
	Enabled   bool
}

func Load() *Config {
	corsOrigins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:4000,http://localhost:3000")
	origins := []string{}
	if corsOrigins != "" {
		// Split comma-separated origins
		for _, origin := range strings.Split(corsOrigins, ",") {
			if trimmed := strings.TrimSpace(origin); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}
	// Default origins if none specified
	if len(origins) == 0 {
		origins = []string{"http://localhost:4000", "http://localhost:3000"}
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "vsq_user"),
			Password: getEnv("DB_PASSWORD", "vsq_password"),
			Name:     getEnv("DB_NAME", "vsq_manpower"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Port:          getEnv("PORT", "8080"),
		SessionSecret: getEnv("SESSION_SECRET", "change-me-in-production"),
		CORS: CORSConfig{
			AllowedOrigins: origins,
		},
		MCP: MCPConfig{
			ServerURL: getEnv("MCP_SERVER_URL", ""),
			APIKey:    getEnv("MCP_API_KEY", ""),
			Enabled:   getEnv("MCP_ENABLED", "false") == "true",
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}


