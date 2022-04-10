package main

import (
	"time"

	"github.com/caarlos0/env/v6"
	zlog "github.com/rs/zerolog/log"
)

type Config struct {
	GRPCServer GRPCServer
}

type GRPCServer struct {
	Address string `env:"GRPC_SERVER_ADDRESS" envDefault:"localhost:9000"`

	MaxRetries   uint          `env:"GRPC_MAX_RETRIES" envDefault:"5"`
	RetryTimeout time.Duration `env:"GRPC_RETRY_TIMEOUT" envDefault:"3s"`
}

func ReadConfig() Config {
	var conf Config
	err := env.Parse(&conf)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to read the config")
	}

	return conf
}
