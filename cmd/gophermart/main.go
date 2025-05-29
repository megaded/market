package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/megaded/market/internal/config"
	"github.com/megaded/market/internal/logger"
	"github.com/megaded/market/internal/manager"
	"github.com/megaded/market/internal/server"
	"github.com/megaded/market/internal/storage"
	"github.com/pressly/goose"
	"go.uber.org/zap"
)

const (
	internalSecond = 15
	workerCount    = 5
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		cancel()
	}()
	logger.SetupLogger("Info")
	cfg := config.GetConfig()
	db, err := sql.Open("postgres", cfg.DBConnString)
	if err != nil {
		logger.Log.Fatal("failed to connect to database", zap.Error(err))
	}
	err = db.Ping()
	if err != nil {
		logger.Log.Fatal("failed ping to database", zap.Error(err))
	}
	err = migrate(db)
	if err != nil {
		logger.Log.Fatal("failed migrate", zap.Error(err))
	}
	storage := storage.NewStorage(cfg)
	client := manager.CreateAccrualClient(cfg.AccrualSystemAddress)
	accrualProcessor := manager.NewAccrualProcessor(storage, &client)
	go accrualProcessor.Run(ctx, internalSecond, workerCount)

	s := server.CreateServer(ctx, cfg, storage)
	if err := s.Start(ctx); err != nil {
		logger.Log.Fatal("server failed", zap.Error(err))
	}
}

func migrate(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	if err := goose.Up(db, "migrations"); err != nil {
		return err
	}

	return nil
}
