package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type (
	Config struct {
		HTTP           HTTP
		PG             PG
		Log            Log
		CancelerWorker CancelerWorker
		NotifierWorker NotifierWorker
		SMTP           SMTP
		Swagger        Swagger
	}

	HTTP struct {
		Port           string `env:"HTTP_PORT,required"`
		UsePreforkMode bool   `env:"HTTP_USE_PREFORK_MODE" envDefault:"false"`
	}

	PG struct {
		PoolMax int    `env:"PG_POOL_MAX,required"`
		URL     string `env:"PG_URL,required"`
	}

	Log struct {
		Level string `env:"LOG_LEVEL,required"`
	}

	CancelerWorker struct {
		BatchSize       int64         `env:"CANCELER_WORKER_BATCH_SIZE" envDefault:"100"`
		PollInterval    time.Duration `env:"CANCELER_WORKER_POLL_INTERVAL" envDefault:"10s"`
		ExpireOpTimeout time.Duration `env:"CANCELER_WORKER_EXPIRE_OPERATION_TIMEOUT" envDefault:"8s"`
		ShutdownTimeout time.Duration `env:"CANCELER_WORKER_SHUTDOWN_TIMEOUT" envDefault:"10s"`
	}

	NotifierWorker struct {
		BatchSize          int64         `env:"NOTIFIER_WORKER_BATCH_SIZE" envDefault:"100"`
		Workers            int           `env:"NOTIFIER_WORKER_WORKERS" envDefault:"3"`
		PollInterval       time.Duration `env:"NOTIFIER_WORKER_POLL_INTERVAL" envDefault:"20s"`
		GetMessagesTimeout time.Duration `env:"NOTIFIER_WORKER_GET_MESSAGES_TIMEOUT" envDefault:"5s"`
		ShutdownTimeout    time.Duration `env:"NOTIFIER_WORKER_SHUTDOWN_TIMEOUT" envDefault:"10s"`
	}

	SMTP struct {
		Username string `env:"SMTP_USERNAME,required"`
		Password string `env:"SMTP_PASSWORD,required"`
		Host     string `env:"SMTP_HOST,required"`
		Port     string `env:"SMTP_PORT,required"`
	}

	Swagger struct {
		Enabled bool `env:"SWAGGER_ENABLED" envDefault:"false"`
	}
)

func New() (*Config, error) {
	cfg := &Config{}

	err := env.Parse(cfg)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	return cfg, nil
}
