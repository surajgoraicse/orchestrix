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
	RestServerPort int
	AppMode        string
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
		RestServerPort: dotenv.GetEnvNumber("REST_SERVER_PORT"),
		AppMode:        dotenv.GetEnv("APP_MODE"),
	}
}
