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
	"github.com/megaded/market/internal/storage/models"
)

type Storager interface {
	GetOrders(ctx context.Context, userID uint) ([]models.Order, error)
	GetOrder(ctx context.Context, orderNumber string) (models.Order, error)
	CreateOrder(ctx context.Context, userID uint, orderNumber string) (models.Order, error)
	GetBalance(ctx context.Context, userID uint) (models.Balance, error)
	CreateUser(ctx context.Context, login string, hash string) (models.User, error)
	GetUserByOrderNumber(ctx context.Context, orderNumber string) (models.User, error)
	GetUser(ctx context.Context, login string) (models.User, error)
	GetProcessingOrders(ctx context.Context) ([]models.Order, error)
	UpdateOrder(ctx context.Context, number string, status string, accrual float64) (models.Order, error)
	AccrualOrder(ctx context.Context, number string, newBalance float64, accrual float64) error
	Withdraw(ctx context.Context, userID uint, orderNumber string, newBalance float64, amount float64) error
	GetOperations(ctx context.Context, userID uint, operationType string) ([]models.Operation, error)
}

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

func CreateServer(ctx context.Context, config *config.Config, storage Storager) (s Server) {
	server := Server{}
	orderManager := manager.CreateOrderManager(storage)
	userManager := manager.CreateUserManager(storage)
	server.Handler = router.CreateRouter(handler.CreateHandlers(storage, orderManager, userManager))
	server.Address = config.Address
	return server
}
