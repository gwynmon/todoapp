package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	ServerPort  string        `env:"SERVER_PORT" envDefault:":8080"`
	LogLevel    string        `env:"LOG_LEVEL" envDefault:"info"`
	PostgresDSN string        `env:"POSTGRES_DSN" envDefault:"postgres://todouser:changeme@localhost:5433/tododb?sslmode=disable"`
	MongoDSN    string        `env:"MONGO_DSN" envDefault:"mongodb://localhost:27017/tododb"`
	RedisDSN    string        `env:"REDIS_DSN" envDefault:"redis://localhost:6379/0"`
	RabbitMQDSN string        `env:"RABBITMQ_DSN" envDefault:"amqp://guest:guest@localhost:5672/"`
	JWTSecret   string        `env:"JWT_SECRET" envDefault:"dev_secret"`
	JWTExpire   time.Duration `env:"JWT_EXPIRE" envDefault:"24h"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
