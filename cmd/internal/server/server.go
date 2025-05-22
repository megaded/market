package server

import (
	"context"
	"net/http"

	"github.com/megaded/market/cmd/internal/config"
	"github.com/megaded/market/cmd/internal/handler"
	"github.com/megaded/market/cmd/internal/logger"
	"github.com/megaded/market/cmd/internal/manager"
	"github.com/megaded/market/cmd/internal/router"
	"github.com/megaded/market/cmd/internal/storage"
)

type Server struct {
	Handler http.Handler
	Address string
}

func (s *Server) Start(ctx context.Context) {
	server := http.Server{Addr: s.Address, Handler: s.Handler}
	go func() {
		<-ctx.Done()
		server.Shutdown(ctx)
	}()
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func CreateServer(ctx context.Context) (s Server) {
	server := Server{}
	logger.SetupLogger("Info")
	serverConfig := config.GetConfig()
	storage := storage.NewStorage(&serverConfig)
	orderManager := manager.CreateOrderManager(&serverConfig)
	server.Handler = router.CreateRouter(handler.CreateHandlers(storage, orderManager), serverConfig)
	server.Address = serverConfig.Address
	return server
}
