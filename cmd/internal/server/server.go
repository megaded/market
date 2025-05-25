package server

import (
	"context"
	"net/http"

	"github.com/megaded/market/cmd/internal/config"
	"github.com/megaded/market/cmd/internal/handler"
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

func CreateServer(ctx context.Context, config config.Config, storage storage.Storager) (s Server) {
	server := Server{}
	orderManager := manager.CreateOrderManager(storage)
	server.Handler = router.CreateRouter(handler.CreateHandlers(storage, orderManager), config)
	server.Address = config.Address
	return server
}
