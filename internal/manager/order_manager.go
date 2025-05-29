package manager

import (
	"context"
	"errors"

	internal_error "github.com/megaded/market/internal/error"
	"github.com/megaded/market/internal/storage/models"
)

type OrderStorager interface {
	GetOrder(ctx context.Context, orderNumber string) (models.Order, error)
	CreateOrder(ctx context.Context, userID uint, orderNumber string) (models.Order, error)
	AccrualOrder(ctx context.Context, number string, newBalance uint, accrual uint) error
	GetBalance(ctx context.Context, userID uint) (models.Balance, error)
	Withdraw(ctx context.Context, userID uint, orderNumber string, newBalance uint, amount uint) error
	UpdateOrder(ctx context.Context, number string, status string, accrual uint) (models.Order, error)
	GetUserByOrderNumber(ctx context.Context, orderNumber string) (models.User, error)
}

type OrderManager struct {
	storage OrderStorager
}

func CreateOrderManager(storage OrderStorager) OrderManager {
	return OrderManager{storage: storage}
}

func (m *OrderManager) UpdateOrder(number string, status string, accrual uint) error {
	err := m.UpdateOrder(number, status, 0)
	return err
}

func (m *OrderManager) AddOrder(ctx context.Context, userID uint, orderNumber string) error {
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

func (m *OrderManager) AccrualOrder(ctx context.Context, number string, accrual uint) error {
	user, err := m.storage.GetUserByOrderNumber(ctx, number)
	if err != nil {
		return err
	}
	balance, err := m.storage.GetBalance(ctx, user.ID)
	if err != nil {
		return err
	}
	newBalance := balance.Balance + accrual
	return m.storage.AccrualOrder(ctx, number, newBalance, accrual)
}

func (m *OrderManager) WithdrawOrder(ctx context.Context, userID uint, number string, withdraw uint) error {
	balance, err := m.storage.GetBalance(ctx, userID)
	if err != nil {
		return err
	}
	if balance.Balance < withdraw {
		return internal_error.ErrInvalidWithdrawSum
	}

	err = m.storage.Withdraw(ctx, userID, number, balance.Balance-withdraw, balance.Withdrawn+withdraw)
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
