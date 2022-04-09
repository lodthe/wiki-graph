package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lodthe/wiki-graph/internal/pathtask"
	"github.com/lodthe/wiki-graph/internal/taskqueue"
	"github.com/lodthe/wiki-graph/internal/wikibfs"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/wagslane/go-rabbitmq"
)

func main() {
	conf := ReadConfig()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	_, cancel := context.WithCancel(context.Background())
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	db, err := setupDatabaseConnection(conf.DB)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to setup database connection")
	}
	defer db.Close()

	rabbitConsumer, err := rabbitmq.NewConsumer(
		conf.AMQP.ConnectionURL,
		rabbitmq.Config{},
		rabbitmq.WithConsumerOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbitConsumer.Close()

	repo := pathtask.NewRepository(db)
	consumer := taskqueue.NewConsumer(rabbitConsumer, conf.AMQP.QueueName, conf.AMQP.RoutingKey)
	handler := wikibfs.NewHandler(repo)

	go func() {
		err := consumer.StartConsuming(handler.HandleTask)
		if err != nil {
			zlog.Fatal().Err(err).Msg("consumer failed")
		}
	}()

	zlog.Info().Msg("consumer has been started")

	<-stop
	cancel()
}

func setupDatabaseConnection(config DB) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", config.PostgresDSN)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(config.MaxConnectionLifetime)
	db.SetMaxOpenConns(config.MaxOpenConnections)
	db.SetMaxIdleConns(config.MaxIdleConnections)

	return db, nil
}
