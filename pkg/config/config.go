package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	MongoDBURI                  string
	MongoDBDatabase             string
	RabbitMQURI                 string
	RabbitMQMenuParsingQueue    string
	RabbitMQProductStatusQueue  string
	RabbitMQDLQQueue            string
	GoogleSheetsCredentialsPath string
	APIPort                     string
	APIHost                     string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	return &Config{
		MongoDBURI:                  getEnv("MONGODB_URI", "mongodb://mongodb:27017"),
		MongoDBDatabase:             getEnv("MONGODB_DATABASE", "menu_parser"),
		RabbitMQURI:                 getEnv("RABBITMQ_URI", "amqp://guest:guest@rabbitmq:5672/"),
		RabbitMQMenuParsingQueue:    getEnv("RABBITMQ_MENU_PARSING_QUEUE", "menu-parsing"),
		RabbitMQProductStatusQueue:  getEnv("RABBITMQ_PRODUCT_STATUS_QUEUE", "product-status"),
		RabbitMQDLQQueue:            getEnv("RABBITMQ_DLQ_QUEUE", "dlq"),
		GoogleSheetsCredentialsPath: getEnv("GOOGLE_SHEETS_CREDENTIALS_PATH", "/app/credentials/credentials.json"),
		APIPort:                     getEnv("API_PORT", "8080"),
		APIHost:                     getEnv("API_HOST", "0.0.0.0"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
