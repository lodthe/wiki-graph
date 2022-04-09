package main

import (
	"time"

	"github.com/caarlos0/env/v6"
	zlog "github.com/rs/zerolog/log"
)

type Config struct {
	DB   DB
	AMQP AMQP
}

type DB struct {
	PostgresDSN string `env:"DB_POSTGRES_DSN,required" envDefault:"host=localhost port=5435 user=user password=password dbname=mafiagraph sslmode=disable"`

	MaxOpenConnections    int           `env:"DB_MAX_OPEN_CONNECTIONS" envDefault:"10"`
	MaxIdleConnections    int           `env:"DB_MAX_IDLE_CONNECTIONS" envDefault:"5"`
	MaxConnectionLifetime time.Duration `env:"DB_MAX_CONNECTION_LIFETIME" envDefault:"5m"`
}

type AMQP struct {
	ConnectionURL string `env:"AMQP_CONNECTION_URL" envDefault:"amqp://user:pass@localhost"`

	QueueName  string `env:"AMQP_QUEUE_NAME" envDefault:"wikigraph_tasks"`
	RoutingKey string `env:"AMQP_ROUTING_KEY" envDefault:"task"`
}

func ReadConfig() Config {
	var conf Config
	err := env.Parse(&conf)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to read the config")
	}

	return conf
}
