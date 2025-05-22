package manager

import (
	"errors"

	"github.com/megaded/market/cmd/internal/config"
	internal_error "github.com/megaded/market/cmd/internal/error"
	"github.com/megaded/market/cmd/internal/storage"
)

type OrderManager struct {
	storage storage.Storager
}

func CreateOrderManager(c *config.Config) OrderManager {
	return OrderManager{storage: storage.NewStorage(c)}
}

func (m *OrderManager) AddOrder(userID int64, orderNumber string) error {
	order, err := m.storage.GetOrder(orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, internal_error.ErrOrderNotFound):
			_, err = m.storage.CreateOrder(userID, orderNumber)
			return err
		default:
			return err
		}
	}
	if order.UserID != uint(userID) {
		return internal_error.ErrOrderAlreadyExistsForAnotherUser
	}
	if order.UserID == uint(userID) {
		return internal_error.ErrOrderAlreadyExists
	}
	return nil
}
