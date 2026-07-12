package config

import "github.com/surajgoraicse/orchestrix/libs/go-libs/dotenv"

type Config struct {
	ServerPort int
}

func NewConfig() (*Config, error) {
	return &Config{
		ServerPort: dotenv.GetEnvNumberOrDefault("SERVER_PORT", 50051),
	}, nil
}
