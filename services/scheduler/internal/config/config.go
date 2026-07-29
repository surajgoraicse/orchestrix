package config

import "github.com/surajgoraicse/orchestrix/libs/go-libs/dotenv"

type Config struct {
	// database
	DbHost     string
	DbPort     int
	DbUser     string
	DbPassword string
	DbName     string
	DbSchema   string
	SSLMode    string

	// server
	ServiceName    string
	RestServerPort int
	AppMode        string
	Environment    string
}

func NewConfig() *Config {
	return &Config{
		// database
		DbHost:     dotenv.GetEnv("DB_HOST"),
		DbPort:     dotenv.GetEnvNumber("DB_PORT"),
		DbUser:     dotenv.GetEnv("DB_USER"),
		DbPassword: dotenv.GetEnv("DB_PASSWORD"),
		DbName:     dotenv.GetEnv("DB_NAME"),
		DbSchema:   dotenv.GetEnv("DB_SCHEMA"),
		SSLMode:    dotenv.GetEnv("SSL_MODE"),

		// server
		ServiceName:    dotenv.GetEnv("SERVICE_NAME"),
		RestServerPort: dotenv.GetEnvNumber("REST_SERVER_PORT"),
		AppMode:        dotenv.GetEnvOrDefault("APP_MODE", "rest"),
		Environment:    dotenv.GetEnvOrDefault("ENVIRONMENT", "dev"),
	}
}
