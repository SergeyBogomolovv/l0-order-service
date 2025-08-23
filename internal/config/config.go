package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type Config struct {
	Env  string `validate:"required,oneof=development stage production"`
	Http Http

	Cors CORS `validate:"required"`

	Kafka Kafka `validate:"required"`

	Postgres Postgres `validate:"required"`
}

type Http struct {
	Host string `validate:"required,hostname|ip"`
	Port string `validate:"required,gt=0,lte=65535"`
}

type Kafka struct {
	GroupID string   `validate:"required"`
	Brokers []string `validate:"required,min=1,dive,hostname_port"`
	Topic   string   `validate:"required"`

	ReaderMaxWait time.Duration `validate:"gte=0"`
	BatchTimeout  time.Duration `validate:"gte=0"`
}

type Postgres struct {
	Host     string `validate:"required,hostname|ip"`
	Port     int    `validate:"required,gt=0,lte=65535"`
	DBName   string `validate:"required"`
	User     string `validate:"required"`
	Password string `validate:"required"`

	SSLMode string `validate:"required,oneof=disable require verify-ca verify-full"`

	MaxOpenConns    int           `validate:"gte=1"`
	MaxIdleConns    int           `validate:"gte=0"`
	ConnMaxLifetime time.Duration `validate:"gte=0"`
}

type CORS struct {
	AllowedOrigins []string `validate:"required,min=1,dive,url"`
}

func New() Config {
	return Config{
		Env: env("ENV", "dev"),

		Http: Http{
			Host: env("HOST", "localhost"),
			Port: env("PORT", "8080"),
		},

		Cors: CORS{
			AllowedOrigins: strings.Split(env("ALLOWED_CORS_ORIGINS", "http://localhost:3000"), ","),
		},

		Kafka: Kafka{
			GroupID: env("KAFKA_GROUP_ID", "order-service"),
			Topic:   env("KAFKA_TOPIC", "orders"),
			Brokers: strings.Split(env("KAFKA_BROKERS", "localhost:9092"), ","),

			ReaderMaxWait: envDuration("KAFKA_READER_MAX_WAIT", 10*time.Millisecond),
			BatchTimeout:  envDuration("KAFKA_BATCH_TIMEOUT", 10*time.Millisecond),
		},

		Postgres: Postgres{
			Port:     envInt("POSTGRES_PORT", 5432),
			Host:     env("POSTGRES_HOST", "localhost"),
			DBName:   env("POSTGRES_DB", "orders"),
			User:     env("POSTGRES_USER", ""),
			Password: env("POSTGRES_PASSWORD", ""),

			SSLMode: env("POSTGRES_SSL_MODE", "disable"),

			MaxOpenConns:    envInt("POSTGRES_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    envInt("POSTGRES_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: envDuration("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
		},
	}
}

func (c Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

func env(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	if len(fallback) == 0 {
		return ""
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if value, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(value)
		if err == nil {
			return d
		}
	}
	return fallback
}
