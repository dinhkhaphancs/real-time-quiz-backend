package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Redis    RedisConfig
	JWT      JWTConfig
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// PostgresConfig represents PostgreSQL database configuration
type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// JWTConfig represents JWT authentication configuration
type JWTConfig struct {
	Secret           string        `mapstructure:"secret"`
	ExpirationTime   time.Duration `mapstructure:"expiration_time"`
	RefreshSecret    string        `mapstructure:"refresh_secret"`
	RefreshExpTime   time.Duration `mapstructure:"refresh_expiration_time"`
	SigningAlgorithm string        `mapstructure:"signing_algorithm"`
	Issuer           string        `mapstructure:"issuer"`
}

// LoadConfig loads configuration from various sources in the following order of precedence:
// 1. Environment variables (with or without APP_ prefix, highest priority)
// 2. Config file specified by APP_CONFIG_FILE environment variable
func LoadConfig() (*Config, error) {
	config := &Config{}
	v := viper.New()

	// Set up environment variables
	v.SetEnvPrefix("APP") // This will prefix all env vars with APP_
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv() // Read environment variables that match

	// Also support standard environment variables without the prefix
	// These take precedence over the prefixed variables
	bindEnvVariables(v)

	// Look for config file
	configFile := getConfigFile()
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			log.Printf("Warning: Unable to read config file: %v", err)
			// Non-fatal error, continue with defaults and env vars
		} else {
			log.Printf("Using config file: %s", v.ConfigFileUsed())
		}
	}

	// Unmarshal the config into our struct
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return config, nil
}

// bindEnvVariables explicitly binds commonly used environment variables
// to their respective config keys for better compatibility
func bindEnvVariables(v *viper.Viper) {
	// Bind standard environment variables (without APP_ prefix)
	v.BindEnv("server.port", "SERVER_PORT")
	v.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	v.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	v.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")

	// PostgreSQL environment variables
	v.BindEnv("postgres.host", "POSTGRES_HOST")
	v.BindEnv("postgres.port", "POSTGRES_PORT")
	v.BindEnv("postgres.user", "POSTGRES_USER")
	v.BindEnv("postgres.password", "POSTGRES_PASSWORD")
	v.BindEnv("postgres.database", "POSTGRES_DB")
	v.BindEnv("postgres.sslmode", "POSTGRES_SSLMODE")

	// Redis environment variables
	v.BindEnv("redis.host", "REDIS_HOST")
	v.BindEnv("redis.port", "REDIS_PORT")
	v.BindEnv("redis.password", "REDIS_PASSWORD")
	v.BindEnv("redis.db", "REDIS_DB")

	// JWT environment variables
	v.BindEnv("jwt.secret", "JWT_SECRET")
	v.BindEnv("jwt.expiration_time", "JWT_EXPIRATION_TIME")
	v.BindEnv("jwt.refresh_secret", "JWT_REFRESH_SECRET")
	v.BindEnv("jwt.refresh_expiration_time", "JWT_REFRESH_EXPIRATION_TIME")
	v.BindEnv("jwt.signing_algorithm", "JWT_SIGNING_ALGORITHM")
	v.BindEnv("jwt.issuer", "JWT_ISSUER")
}

// getConfigFile returns the config file path from APP_CONFIG_FILE environment variable
func getConfigFile() string {
	// Only check environment variable for config file path
	if configPath := os.Getenv("APP_CONFIG_FILE"); configPath != "" {
		return configPath
	}

	return "" // No config file specified
}

// GetConnectionString returns a formatted PostgreSQL connection string
func (p PostgresConfig) GetConnectionString() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.Database, p.SSLMode)
}

// GetAddr returns Redis address in the format "host:port"
func (r RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
