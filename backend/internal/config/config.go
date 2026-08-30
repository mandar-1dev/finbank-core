// Package config loads runtime configuration from environment variables.
// CoreBank has no login system, so the only secrets here are database
// credentials for a purely local, fictional simulation database.
package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	ServerPort string
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Load reads configuration from the environment, falling back to sane
// local-development defaults so the project runs out of the box.
func Load() Config {
	return Config{
		DBHost:     getEnv("DB_HOST", "127.0.0.1"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "corebank_db"),
		DBUser:     getEnv("DB_USERNAME", "corebank"),
		DBPassword: getEnv("DB_PASSWORD", "corebank_pass"),
		ServerPort: getEnv("SERVER_PORT", "8080"),
	}
}

// DSN builds the MySQL data source name for go-sql-driver/mysql.
func (c Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=Local",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}
