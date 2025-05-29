package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/megaded/market/internal/config"
	"github.com/megaded/market/internal/handler"
	"github.com/megaded/market/internal/manager"
	"github.com/megaded/market/internal/router"
	"github.com/megaded/market/internal/storage"
)

type Server struct {
	Handler http.Handler
	Address string
}

func (s *Server) Start(ctx context.Context) error {
	server := http.Server{Addr: s.Address, Handler: s.Handler}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func CreateServer(ctx context.Context, config *config.Config, storage storage.Storager) (s Server) {
	server := Server{}
	orderManager := manager.CreateOrderManager(storage)
	userManager := manager.CreateUserManager(storage)
	server.Handler = router.CreateRouter(handler.CreateHandlers(storage, orderManager, userManager))
	server.Address = config.Address
	return server
}
