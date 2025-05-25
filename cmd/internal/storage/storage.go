package storage

import (
	"errors"

	"github.com/megaded/market/cmd/internal/config"
	internal_error "github.com/megaded/market/cmd/internal/error"
	"github.com/megaded/market/cmd/internal/identity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storager interface {
	GetOrders(userID int64) ([]Order, error)
	GetOrder(orderNumber string) (Order, error)
	CreateOrder(userID int64, orderNumber string) (Order, error)
	GetBalance(userID int64) (Balance, error)
	CreateUser(login string, hash string) (User, error)
	GetUser(login string) (User, error)
	GetProcessingOrders() ([]Order, error)
	UpdateOrder(number string, status string, accrual int) error
	UpdateBalance(userID int, amount int) error
	CreateOperation(userID int, orderID int, value int) error
}

type storage struct {
	db       *gorm.DB
	identity identity.IdentityProvider
}

func (s *storage) CreateOperation(userID int, orderID int, value int) error {
	operation := Operation{UserID: uint(userID), OrderID: uint(orderID), Value: int64(value)}
	r := s.db.Create(&operation)
	return r.Error
}

func (s *storage) UpdateBalance(userID int, amount int) error {
	if err := s.db.Model(Balance{}).Where("user_id = ?", userID).Select("amount").Updates(map[string]interface{}{"amount": amount}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return internal_error.ErrOrderNotFound
		}
		return err
	}
	return nil
}

func (s *storage) CreateOrder(userID int64, orderNumber string) (Order, error) {
	order := Order{UserID: uint(userID), Number: orderNumber}
	r := s.db.Create(&order)
	return order, r.Error
}
func (s *storage) UpdateOrder(number string, status string, accrual int) error {
	if err := s.db.Model(Order{}).Where("number = ?", number).Select("status", "accrual").Updates(map[string]interface{}{"status": status, "accrual": accrual}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return internal_error.ErrOrderNotFound
		}
		return err
	}
	return nil
}

func (s *storage) GetProcessingOrders() ([]Order, error) {
	var orders []Order
	result := s.db.Where("status = ? or status = ?", OrderStatusNew, OrderStatusProcessing).Find(&orders)
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return orders, internal_error.ErrOrderNotFound
	default:
		return orders, result.Error
	}
}

func (s *storage) GetOrders(userID int64) ([]Order, error) {
	var orders []Order
	result := s.db.Where("user_id = ?", userID).Find(&orders)
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return orders, internal_error.ErrOrderNotFound
	default:
		return orders, result.Error
	}
}

func (s *storage) GetOrder(orderNumber string) (Order, error) {
	var order Order
	result := s.db.Where("number = ?", orderNumber).First(&order)
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return order, internal_error.ErrOrderNotFound
	default:
		return order, result.Error
	}
}

func (s *storage) GetUser(login string) (User, error) {
	var user User
	result := s.db.Where("name = ?", login).First(&user)
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return user, internal_error.ErrUserNotFound
	default:
		return user, result.Error
	}
}

func (s *storage) CreateUser(login string, password string) (User, error) {
	if login == "" || password == "" {
		return User{}, internal_error.ErrEmptyLoginOrPassword
	}
	var user User
	result := s.db.Where("name = ?", login).First(&user)
	switch {
	case result.Error == nil:
		return User{}, internal_error.ErrUserAlreadyExists
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		r := s.db.Create(&User{Name: login, Hash: s.identity.HashPassword(password)})
		if r.Error != nil {
			return user, r.Error
		}
		balance := Balance{UserID: user.ID}
		r = s.db.Create(&balance)
		return user, r.Error
	default:
		return User{}, result.Error

	}
}

func (s *storage) GetBalance(userId int64) (Balance, error) {
	var balance Balance
	result := s.db.Where("user_id = ?", userId).First(&balance)
	return balance, result.Error
}

func NewStorage(c *config.Config) Storager {
	db, err := gorm.Open(postgres.Open(c.DBConnString), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Order{})
	db.AutoMigrate(&Balance{})
	db.AutoMigrate(&Operation{})
	return &storage{db: db}
}
