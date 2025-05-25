package storage

import (
	"context"
	"errors"
	"math"

	"github.com/megaded/market/cmd/internal/config"
	internal_error "github.com/megaded/market/cmd/internal/error"
	"github.com/megaded/market/cmd/internal/identity"
	"github.com/megaded/market/cmd/internal/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storager interface {
	GetOrders(ctx context.Context, userID int64) ([]Order, error)
	GetOrder(ctx context.Context, orderNumber string) (Order, error)
	CreateOrder(ctx context.Context, userID int64, orderNumber string) (Order, error)
	GetBalance(ctx context.Context, userID int64) (Balance, error)
	CreateUser(ctx context.Context, login string, hash string) (User, error)
	GetUser(ctx context.Context, login string) (User, error)
	GetProcessingOrders(ctx context.Context) ([]Order, error)
	UpdateOrder(ctx context.Context, number string, status string, accrual float64) error
	Accrual(ctx context.Context, userID int, orderNumber string, amount float64) error
	Withdraw(ctx context.Context, userID int, orderNumber string, amount float64) error
	CreateOperation(ctx context.Context, userID int, orderNumber string, operationType string, value float64) error
	GetOperations(ctx context.Context, userID int, operationType string) ([]Operation, error)
}

type storage struct {
	db       *gorm.DB
	identity identity.IdentityProvider
}

func (s *storage) GetOperations(ctx context.Context, userID int, operationType string) ([]Operation, error) {
	var operations []Operation
	result := s.db.WithContext(ctx).Model(&Operation{}).Where("user_id = ? AND operation_type = ?", userID, operationType).Find(&operations)
	switch {
	case len(operations) == 0:
		return operations, internal_error.ErrWithdrawalNotFound
	default:
		if result.Error != nil {
			logger.Log.Info(result.Error.Error())
		}
		return operations, result.Error
	}
}

func (s *storage) CreateOperation(ctx context.Context, userID int, orderNumber string, operationType string, value float64) error {
	operation := Operation{UserID: uint(userID), Order: orderNumber, Value: value, OperationType: operationType}
	r := s.db.WithContext(ctx).Create(&operation)
	return r.Error
}

func (s *storage) Withdraw(ctx context.Context, userID int, orderNumber string, amount float64) error {
	amount = math.Round(amount*100) / 100
	db := s.db.WithContext(ctx)
	db.Begin()
	defer db.Commit()
	var balance Balance
	r := db.Where("user_id = ?", userID).First(&balance)
	if r.Error != nil {
		return r.Error
	}
	newBalance := math.Round(balance.Balance-amount*100) / 100
	if err := db.Model(Balance{}).Where("user_id = ?", userID).Select("balance", "withdrawn").Updates(map[string]interface{}{"balance": newBalance, "withdrawn": balance.Withdrawn + amount}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return internal_error.ErrOrderNotFound
		}
		db.Rollback()
		return err
	}
	err := s.CreateOperation(ctx, userID, orderNumber, string(Withdraw), amount)
	if err != nil {
		db.Rollback()
	}
	return nil
}

func (s *storage) Accrual(ctx context.Context, userID int, orderNumber string, amount float64) error {
	db := s.db.WithContext(ctx)
	defer db.Commit()
	var balance Balance
	r := db.Where("user_id = ?", userID).First(&balance)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return internal_error.ErrOrderNotFound
		}
		return r.Error
	}
	newBalance := math.Round(balance.Balance+amount*100) / 100
	r = db.Model(Balance{}).Where("user_id = ?", userID).Select("balance").Updates(map[string]interface{}{"balance": newBalance})
	if r.Error != nil {
		return r.Error
	}
	err := s.CreateOperation(ctx, userID, orderNumber, string(Accrual), amount)
	if err != nil {
		db.Rollback()
	}
	return nil
}

func (s *storage) CreateOrder(ctx context.Context, userID int64, orderNumber string) (Order, error) {
	order := Order{UserID: uint(userID), Number: orderNumber, Status: OrderStatusNew}
	r := s.db.WithContext(ctx).Create(&order)
	return order, r.Error
}
func (s *storage) UpdateOrder(ctx context.Context, number string, status string, accrual float64) error {
	if err := s.db.WithContext(ctx).Model(Order{}).Where("number = ?", number).Select("status", "accrual").Updates(map[string]interface{}{"status": status, "accrual": accrual}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return internal_error.ErrOrderNotFound
		}
		return err
	}
	return nil
}

func (s *storage) GetProcessingOrders(ctx context.Context) ([]Order, error) {
	var orders []Order
	result := s.db.WithContext(ctx).Where("status = ? or status = ?", OrderStatusNew, OrderStatusProcessing).Find(&orders)
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return orders, internal_error.ErrOrderNotFound
	default:
		return orders, result.Error
	}
}

func (s *storage) GetOrders(ctx context.Context, userID int64) ([]Order, error) {
	var orders []Order
	result := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&orders)
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return orders, internal_error.ErrOrderNotFound
	default:
		return orders, result.Error
	}
}

func (s *storage) GetOrder(ctx context.Context, orderNumber string) (Order, error) {
	var order Order
	result := s.db.WithContext(ctx).Where("number = ?", orderNumber).First(&order)
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return order, internal_error.ErrOrderNotFound
	default:
		return order, result.Error
	}
}

func (s *storage) GetUser(ctx context.Context, login string) (User, error) {
	var user User
	result := s.db.WithContext(ctx).Where("name = ?", login).First(&user)
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return user, internal_error.ErrUserNotFound
	default:
		return user, result.Error
	}
}

func (s *storage) CreateUser(ctx context.Context, login string, password string) (User, error) {
	db := s.db.WithContext(ctx)
	defer db.Commit()
	if login == "" || password == "" {
		return User{}, internal_error.ErrEmptyLoginOrPassword
	}
	var user User
	result := db.Where("name = ?", login).First(&user)
	switch {
	case result.Error == nil:
		return User{}, internal_error.ErrUserAlreadyExists
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		newUser := User{
			Name: login, Hash: s.identity.HashPassword(password),
		}
		r := db.Create(&newUser)
		if r.Error != nil {
			db.Rollback()
			return user, r.Error
		}
		balance := Balance{UserID: newUser.ID}
		r = db.Create(&balance)
		if r.Error != nil {
			db.Rollback()
		}
		user = newUser
		return user, r.Error
	default:
		return User{}, result.Error

	}
}

func (s *storage) GetBalance(ctx context.Context, userID int64) (Balance, error) {
	var balance Balance
	result := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&balance)
	if result.Error != nil {
		logger.Log.Info(result.Error.Error())
	}
	return balance, result.Error
}

func NewStorage(c *config.Config) Storager {
	db, err := gorm.Open(postgres.Open(c.DBConnString), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	/*db.AutoMigrate(&User{})
	db.AutoMigrate(&Order{})
	db.AutoMigrate(&Balance{})
	db.AutoMigrate(&Operation{})*/
	return &storage{db: db}
}
