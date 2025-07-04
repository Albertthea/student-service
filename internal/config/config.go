// Package config provides functionality for loading configuration
// from a YAML file and environment variables.
package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// Config describes the structure of the server and database configuration.
type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`

	PostgreSQL struct {
		Host          string `yaml:"host"`
		Port          int    `yaml:"port"`
		Authorisation struct {
			Env struct {
				LoginEnv    string `yaml:"login"`
				PasswordEnv string `yaml:"password"`
			} `yaml:"env"`
		} `yaml:"authorisation"`
	} `yaml:"postgresql"`

	DBLogin    string
	DBPassword string
}

// LoadConfig loads configuration from the specified YAML file
// and substitutes authorization parameters from environment variables.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.DBLogin = os.Getenv(cfg.PostgreSQL.Authorisation.Env.LoginEnv)
	cfg.DBPassword = os.Getenv(cfg.PostgreSQL.Authorisation.Env.PasswordEnv)

	if cfg.DBLogin == "" || cfg.DBPassword == "" {
		log.Println("Warning: DB credentials are missing")
		return nil, fmt.Errorf("database credentials not found in env variables: %s, %s",
			cfg.PostgreSQL.Authorisation.Env.LoginEnv,
			cfg.PostgreSQL.Authorisation.Env.PasswordEnv,
		)
	}

	return &cfg, nil
}
