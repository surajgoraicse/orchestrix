package config

import "github.com/surajgoraicse/orchestrix/packages/go-package/dotenv"

type Config struct {
	DbPort      int
	DbHost      string
	DbUser      string
	DbPassword  string
	DbName      string
	DatabaseURL string
	ServerPort  int
}

func NewConfig() *Config {
	return &Config{
		DbPort:      dotenv.GetEnvNumber("DB_PORT"),
		DbHost:      dotenv.GetEnv("DB_HOST"),
		DbUser:      dotenv.GetEnv("DB_USER"),
		DbPassword:  dotenv.GetEnv("DB_PASSWORD"),
		DbName:      dotenv.GetEnv("DB_NAME"),
		DatabaseURL: dotenv.GetEnv("DATABASE_URL"),
		ServerPort:  dotenv.GetEnvNumber("SCHEDULER_SERVER_PORT"),
	}
}
