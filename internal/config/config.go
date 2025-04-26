package config

import (
	"time"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Redis    RedisConfig
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// PostgresConfig represents PostgreSQL database configuration
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
	SSLMode  string
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// NewConfig creates a default configuration
func NewConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         8080,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Postgres: PostgresConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "postgres",
			Database: "quiz_db",
			SSLMode:  "disable",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
	}
}