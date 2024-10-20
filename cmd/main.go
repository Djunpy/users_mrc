package main

import (
	"context"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
	config2 "users_mrc/config"
	db "users_mrc/db/sqlc"
	"users_mrc/infrastructure/database"
	"users_mrc/infrastructure/server"
	"users_mrc/infrastructure/worker"
	logger2 "users_mrc/logger"
)

var interruptSignals = []os.Signal{
	os.Interrupt,
	syscall.SIGTERM,
	syscall.SIGINT,
}

func main() {
	config, err := config2.LoadConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("cannot load config")
	}
	logger, file := logger2.ConfigureLogger(config.Environment)
	defer file.Close()
	ctx, stop := signal.NotifyContext(context.Background(), interruptSignals...)
	defer stop()
	connPool, err := pgxpool.New(ctx, config.PostgresSource)
	defer connPool.Close()
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot connect to db")
	}
	err = connPool.Ping(ctx)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot connect to db")
	}
	err = database.RunDBMigration(config.MigrationURL, config.PostgresSource)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot run migration")
	}
	store := db.NewStore(connPool)
	redisOpt := asynq.RedisClientOpt{Addr: config.RedisAddress}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)
	waitGroup, ctx := errgroup.WithContext(ctx)

	err = worker.RunTaskProcessor(ctx, waitGroup, config, redisOpt, store, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot run task processor")
	}

	err = server.RunGinServer(config, store, taskDistributor)
	if err != nil {
		logger.Fatal().Err(err).Msg("cannot run gin server")
	}

	err = waitGroup.Wait()
	if err != nil {
		logger.Fatal().Err(err).Msg("error from wait group")
	}
}
