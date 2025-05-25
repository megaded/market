package manager

import (
	"errors"

	internal_error "github.com/megaded/market/cmd/internal/error"
	"github.com/megaded/market/cmd/internal/storage"
)

type OrderManager struct {
	storage storage.Storager
}

func CreateOrderManager(storage storage.Storager) OrderManager {
	return OrderManager{storage: storage}
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

func (m *OrderManager) UpdateOrder(number string, status string, accrual int) error {
	if err := m.storage.UpdateOrder(number, status, accrual); err != nil {
		return err
	}
	order, err := m.storage.GetOrder(number)
	if err != nil {
		return err
	}
	err = m.storage.UpdateBalance(int(order.UserID), accrual)
	return err
}
