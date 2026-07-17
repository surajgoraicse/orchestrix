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
	DbSchema   string
	SSLMode    string

	// coordinator server
	ServerPort        int
	DbScanInterval    time.Duration
	HeartbeatInterval time.Duration
	DeadWorkerTTL     time.Duration
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
		ServerPort:        dotenv.GetEnvNumber("SERVER_PORT"),
		DbScanInterval:    time.Duration(dotenv.GetEnvNumberOrDefault("DB_SCAN_INTERVAL", 10)) * time.Second,
		HeartbeatInterval: time.Duration(dotenv.GetEnvNumberOrDefault("HEARTBEAT_INTERVAL", 10)) * time.Second,
		DeadWorkerTTL:     time.Duration(dotenv.GetEnvNumberOrDefault("DEAD_WORKER_TTL", 20)) * time.Second,
	}
}
