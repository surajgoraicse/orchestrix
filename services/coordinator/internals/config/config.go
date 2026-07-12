package config

import (
	"time"

	"github.com/surajgoraicse/orchestrix/libs/go-libs/dotenv"
)

type Config struct {
	// database
	DbHost     string
	DbPort     int
	DbUser     string
	DbPassword string
	DbName     string
	SSLMode    string

	// coordinator server
	ServerPort         int
	HeartbeatInterval  time.Duration
	MaxHeartbeatMisses int
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
		ServerPort:         dotenv.GetEnvNumber("SERVER_PORT"),
		HeartbeatInterval:  time.Duration(dotenv.GetEnvNumberOrDefault("HEARTBEAT_INTERVAL", 10)) * time.Second,
		MaxHeartbeatMisses: dotenv.GetEnvNumberOrDefault("MAX_HEARTBEAT_MISSES", 3),
	}
}
