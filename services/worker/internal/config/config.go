package config

import (
	"time"

	"github.com/surajgoraicse/orchestrix/libs/go-libs/dotenv"
)

type Config struct {
	CoordinatorAddress string
	HeartbeatInterval  time.Duration
}

func NewConfig() *Config {
	return &Config{
		CoordinatorAddress: dotenv.GetEnv("COORDINATOR_ADDRESS"),
		HeartbeatInterval:  time.Duration(dotenv.GetEnvNumberOrDefault("HEARTBEAT_INTERVAL", 10)) * time.Second,
	}
}
