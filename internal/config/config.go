package config

import (
	"errors"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServerPort  string        `env:"SERVER_PORT" env-default:":8080"`
	PostgresDSN string        `env:"POSTGRES_DSN" env-default:"postgres://todouser:changeme@localhost:5433/tododb?sslmode=disable"`
	JWTSecret   string        `env:"JWT_SECRET" env-default:"dev_secret"`
	JWTExpire   time.Duration `env:"JWT_EXPIRE" env-default:"24h"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	if cfg.PostgresDSN == "" {
		return nil, errors.New("POSTGRES_DSN is required")
	}

	return &cfg, nil
}
