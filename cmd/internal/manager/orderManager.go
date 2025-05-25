package manager

import (
	"context"
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

func (m *OrderManager) AddOrder(ctx context.Context, userID int64, orderNumber string) error {
	if !validateOrderNumber(orderNumber) {
		return internal_error.ErrInvalidOrderNumber
	}
	order, err := m.storage.GetOrder(ctx, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, internal_error.ErrOrderNotFound):
			_, err = m.storage.CreateOrder(ctx, userID, orderNumber)
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

func (m *OrderManager) AccrualOrder(ctx context.Context, number string, status string, accrual float64) error {
	if err := m.storage.UpdateOrder(ctx, number, status, accrual); err != nil {
		return err
	}
	order, err := m.storage.GetOrder(ctx, number)
	if err != nil {
		return err
	}
	err = m.storage.Accrual(ctx, int(order.UserID), number, accrual)
	return err
}

func (m *OrderManager) WithdrawOrder(ctx context.Context, userID int, number string, withdraw float64) error {
	balance, err := m.storage.GetBalance(ctx, int64(userID))
	if err != nil {
		return err
	}
	if balance.Balance < withdraw {
		return internal_error.ErrInvalidWithdrawSum
	}

	err = m.storage.Withdraw(ctx, userID, number, withdraw)
	return err
}

func validateOrderNumber(number string) bool {
	var sum int
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		r := number[i]

		if r < '0' || r > '9' {
			return false
		}

		digit := int(r - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}
