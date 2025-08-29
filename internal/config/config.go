// Package config provides application configuration loading from YAML and environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config defines the root configuration structure of the application.
type Config struct {
	Server ServerConfig `yaml:"server"`
	PG     PostgreSQL   `yaml:"postgresql"`
	Log    LogConfig    `yaml:"log"`
	GRPC   GRPCConfig   `yaml:"grpc"`
}

// ServerConfig holds HTTP/gRPC server settings.
type ServerConfig struct {
	Host                string `yaml:"host"`
	Port                int    `yaml:"port"`
	GracefulShutdownSec int    `yaml:"graceful_shutdown_sec"`
}

// GRPCConfig holds optional gRPC-specific settings.
type GRPCConfig struct {
	EnableReflection bool `yaml:"enable_reflection"`
	EnableHealth     bool `yaml:"enable_health"`
}

// LogConfig defines logging settings.
type LogConfig struct {
	Level string `yaml:"level"`
}

// PostgreSQL defines PostgreSQL database connection settings.
type PostgreSQL struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	DBName  string `yaml:"dbname"`
	SSLMode string `yaml:"sslmode"`

	User     string `yaml:"user"`
	Password string `yaml:"password"`

	AuthEnv struct {
		UserEnv string `yaml:"user_env"`
		PassEnv string `yaml:"pass_env"`
	} `yaml:"auth_env"`
}

// Load reads the configuration from a YAML file, applies defaults and environment overrides.
func Load(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	cfg.applyDefaults()
	cfg.applyEnvOverrides()
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return &cfg, nil
}

// applyDefaults sets default values for configuration fields if they are not provided.
func (c *Config) applyDefaults() {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 50051
	}
	if c.Server.GracefulShutdownSec == 0 {
		c.Server.GracefulShutdownSec = 10
	}
	if c.PG.SSLMode == "" {
		c.PG.SSLMode = "disable"
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
}

// applyEnvOverrides overrides config values using environment variables if present.
func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("APP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.Server.Port = p
		}
	}
	if v := os.Getenv("APP_GRPC_REFLECTION"); v != "" {
		c.GRPC.EnableReflection = parseBool(v, c.GRPC.EnableReflection)
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.Log.Level = v
	}

	if env := strings.TrimSpace(c.PG.AuthEnv.UserEnv); env != "" {
		if v := os.Getenv(env); v != "" {
			c.PG.User = v
		}
	}
	if env := strings.TrimSpace(c.PG.AuthEnv.PassEnv); env != "" {
		if v := os.Getenv(env); v != "" {
			c.PG.Password = v
		}
	}

	if v := os.Getenv("PGHOST"); v != "" {
		c.PG.Host = v
	}
	if v := os.Getenv("PGPORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.PG.Port = p
		}
	}
	if v := os.Getenv("PGDATABASE"); v != "" {
		c.PG.DBName = v
	}
	if v := os.Getenv("PGUSER"); v != "" {
		c.PG.User = v
	}
	if v := os.Getenv("PGPASSWORD"); v != "" {
		c.PG.Password = v
	}
	if v := os.Getenv("PGSSLMODE"); v != "" {
		c.PG.SSLMode = v
	}
}

// validate checks the configuration for required fields and constraints.
func (c *Config) validate() error {
	if c.Server.Port <= 0 {
		return errors.New("server.port must be > 0")
	}
	if c.PG.Host == "" || c.PG.Port == 0 || c.PG.DBName == "" {
		return errors.New("postgresql host/port/dbname must be set")
	}
	if c.PG.User == "" || c.PG.Password == "" {
		if c.PG.AuthEnv.UserEnv == "" || c.PG.AuthEnv.PassEnv == "" {
			return errors.New("postgresql user/password are empty and auth_env.user_env/pass_env not provided")
		}
	}
	return nil
}

// BuildPostgresDSN builds a PostgreSQL DSN string from configuration.
func (c *Config) BuildPostgresDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.PG.Host, c.PG.Port, c.PG.User, c.PG.Password, c.PG.DBName, c.PG.SSLMode,
	)
}

// parseBool parses a boolean value from string with a default fallback.
func parseBool(s string, def bool) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "t", "yes", "y", "on":
		return true
	case "0", "false", "f", "no", "n", "off":
		return false
	default:
		return def
	}
}
