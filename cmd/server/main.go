package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lodthe/wiki-graph/internal/pathtask"
	"github.com/lodthe/wiki-graph/internal/taskqueue"
	"github.com/lodthe/wiki-graph/internal/wikigraphserver"
	"github.com/lodthe/wiki-graph/pkg/wikigraphpb"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/wagslane/go-rabbitmq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
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

	publisher, err := rabbitmq.NewPublisher(
		conf.AMQP.ConnectionURL,
		rabbitmq.Config{},
		rabbitmq.WithPublisherOptionsLogging,
	)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to connect to RabbitMQ")
	}
	defer publisher.Close()

	repo := pathtask.NewRepository(db)
	producer := taskqueue.NewProducer(publisher, conf.AMQP.ExchangeName, conf.AMQP.RoutingKey)
	wikiGraphServer := wikigraphserver.New(repo, producer)

	srv, lis, err := registerServer(conf.GRPCServer, wikiGraphServer)
	if err != nil {
		zlog.Fatal().Err(err).Str("address", conf.GRPCServer.Address).Msg("server registration failed")
	}

	go func() {
		err := srv.Serve(lis)
		if err != nil {
			zlog.Fatal().Err(err).Msg("server failed")
		}
	}()

	zlog.Info().Str("address", conf.GRPCServer.Address).Msg("server started")

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

func registerServer(conf GRPCServer, wikiGraphServer *wikigraphserver.Server) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", conf.Address)
	if err != nil {
		return nil, nil, errors.Wrap(err, "listen failed")
	}

	grpcServer := grpc.NewServer(
		grpc.ConnectionTimeout(conf.Timeout),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: conf.KeepAlive,
			Time:              conf.KeepAlive,
			Timeout:           conf.KeepAlive,
		}),
		StdUnaryMiddleware(),
		StdStreamMiddleware(),
	)

	wikigraphpb.RegisterWikiGraphServer(grpcServer, wikiGraphServer)

	reflection.Register(grpcServer)

	return grpcServer, lis, nil
}
