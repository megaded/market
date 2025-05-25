package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/megaded/market/cmd/internal/config"
	"github.com/megaded/market/cmd/internal/logger"
	"github.com/megaded/market/cmd/internal/manager"
	"github.com/megaded/market/cmd/internal/server"
	"github.com/megaded/market/cmd/internal/storage"
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
	storage := storage.NewStorage(&cfg)
	client := manager.CreateAccrualClient(cfg.AccrualSystemAddress)
	accrualProcessor := manager.NewAccrualProcessor(storage, &client)
	go accrualProcessor.Run(ctx, 15, 5)

	s := server.CreateServer(ctx, cfg, storage)
	s.Start(ctx)
}
