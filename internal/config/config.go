package config

import (
	"os"
)

type Config struct {
	RedisHost        string
	RabbitMQHost     string
	RabbitMQUser     string
	RabbitMQPassword string
	SQLiteDBPath     string
}

func LoadConfig() (*Config, error) {
	redisHost := getEnvOrDefault("REDIS_HOST", "localhost")
	rabbitMQHost := getEnvOrDefault("RABBITMQ_HOST", "localhost")
	rabbitMQUser := getEnvOrDefault("RABBITMQ_USER", "user")
	rabbitMQPassword := getEnvOrDefault("RABBITMQ_PASSWORD", "password")
	sqliteDBPath := getEnvOrDefault("SQLITE_DB_PATH", "../data/app.db")

	return &Config{
		RedisHost:        redisHost,
		RabbitMQHost:     rabbitMQHost,
		RabbitMQUser:     rabbitMQUser,
		RabbitMQPassword: rabbitMQPassword,
		SQLiteDBPath:     sqliteDBPath,
	}, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
