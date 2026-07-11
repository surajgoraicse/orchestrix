package config

import "github.com/surajgoraicse/orchestrix/libs/go-libs/dotenv"

type Config struct {
	// database
	DbHost     string
	DbPort     int
	DbUser     string
	DbPassword string
	DbName     string
	SSLMode    string

	// server
	ServerPort int
}

func NewConfig() *Config {
	return &Config{
		// database
		DbHost:     dotenv.GetEnv("DB_HOST"),
		DbPort:     dotenv.GetEnvNumber("DB_PORT"),
		DbUser:     dotenv.GetEnv("DB_USER"),
		DbPassword: dotenv.GetEnv("DB_PASSWORD"),
		DbName:     dotenv.GetEnv("DB_NAME"),
		SSLMode:    dotenv.GetEnv("SSL_MODE"),

		// server
		ServerPort: dotenv.GetEnvNumber("SERVER_PORT"),
	}
}
