package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	AuthServerPort     string `env:"AUTH_SERVER_PORT" envDefault:":8081"`
	TasksServerPort    string `env:"TASKS_SERVER_PORT" envDefault:":8082"`
	NotifierServerPort string `env:"NOTIFIER_SERVER_PORT" envDefault:":8083"`

	LogLevel    string        `env:"LOG_LEVEL" envDefault:"info"`
	PostgresDSN string        `env:"POSTGRES_DSN" envDefault:"postgres://todouser:changeme@localhost:5433/tododb?sslmode=disable"`
	MongoDSN    string        `env:"MONGO_DSN" envDefault:"mongodb://localhost:27017/tododb"`
	RedisDSN    string        `env:"REDIS_DSN" envDefault:"redis://localhost:6379/0"`
	RabbitMQDSN string        `env:"RABBITMQ_DSN" envDefault:"amqp://guest:guest@localhost:5672/"`
	JWTSecret   string        `env:"JWT_SECRET" envDefault:"dev_secret"`
	JWTExpire   time.Duration `env:"JWT_EXPIRE" envDefault:"24h"`

	TasksServiceURL       string        `env:"TASKS_SERVICE_URL" envDefault:"http://localhost:8082"`
	InternalSecret        string        `env:"INTERNAL_SECRET" envDefault:"internal_dev_secret"`
	DeadlineCheckInterval time.Duration `env:"DEADLINE_CHECK_INTERVAL" envDefault:"1h"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
