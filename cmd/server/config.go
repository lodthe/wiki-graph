package main

import (
	"time"

	"github.com/caarlos0/env/v6"
	zlog "github.com/rs/zerolog/log"
)

type Config struct {
	DB         DB
	GRPCServer GRPCServer
	AMQP       AMQP
}

type DB struct {
	PostgresDSN string `env:"DB_POSTGRES_DSN,required" envDefault:"host=localhost port=5435 user=user password=password dbname=mafiagraph sslmode=disable"`

	MaxOpenConnections    int           `env:"DB_MAX_OPEN_CONNECTIONS" envDefault:"10"`
	MaxIdleConnections    int           `env:"DB_MAX_IDLE_CONNECTIONS" envDefault:"5"`
	MaxConnectionLifetime time.Duration `env:"DB_MAX_CONNECTION_LIFETIME" envDefault:"5m"`
}

type GRPCServer struct {
	Address string `env:"GRPC_SERVER_ADDRESS" envDefault:"0.0.0.0:9000"`

	Timeout   time.Duration `env:"GRPC_SERVER_TIMEOUT" envDefault:"10s"`
	KeepAlive time.Duration `env:"GRPC_SERVER_KEEP_ALIVE" envDefault:"500ms"`
}

type AMQP struct {
	ConnectionURL string `env:"AMQP_CONNECTION_URL" envDefault:"amqp://user:pass@localhost"`

	ExchangeName string `env:"AMQP_EXCHANGE_NAME" envDefault:"wikigraph"`
	RoutingKey   string `env:"AMQP_ROUTING_KEY" envDefault:"task"`
}

func ReadConfig() Config {
	var conf Config
	err := env.Parse(&conf)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to read the config")
	}

	return conf
}
