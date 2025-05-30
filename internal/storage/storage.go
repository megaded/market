package storage

import (
	"context"
	"errors"

	"github.com/megaded/market/internal/config"
	internal_error "github.com/megaded/market/internal/error"
	"github.com/megaded/market/internal/logger"
	"github.com/megaded/market/internal/storage/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type storage struct {
	db *gorm.DB
}

func (s *storage) GetOperations(ctx context.Context, userID uint, operationType string) ([]models.Operation, error) {
	var operations []models.Operation
	result := s.db.WithContext(ctx).Model(&models.Operation{}).Where("user_id = ? AND operation_type = ?", userID, operationType).Find(&operations)
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

func createOperation(db *gorm.DB, userID uint, orderNumber string, operationType string, value float64) error {
	operation := models.Operation{UserID: uint(userID), OrderNumber: orderNumber, Value: value, OperationType: operationType}
	r := db.Create(&operation)
	return r.Error
}

func (s *storage) Withdraw(ctx context.Context, userID uint, orderNumber string, newBalance float64, amount float64) error {
	db := s.db.WithContext(ctx)
	db.Begin()
	defer db.Commit()
	var balance models.Balance
	r := db.Where("user_id = ?", userID).First(&balance)
	if r.Error != nil {
		return r.Error
	}
	if err := db.Model(models.Balance{}).Where("user_id = ?", userID).Select("balance", "withdrawn").Updates(map[string]interface{}{"balance": newBalance, "withdrawn": amount}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return internal_error.ErrOrderNotFound
		}
		db.Rollback()
		return err
	}
	err := createOperation(db, userID, orderNumber, string(models.Withdraw), amount)
	if err != nil {
		db.Rollback()
	}
	return nil
}

func accrualBalance(db *gorm.DB, userID uint, orderNumber string, newBalance float64, amount float64) error {
	var balance models.Balance
	r := db.Where("user_id = ?", userID).First(&balance)
	if r.Error != nil {
		if errors.Is(r.Error, gorm.ErrRecordNotFound) {
			return internal_error.ErrOrderNotFound
		}
		return r.Error
	}
	r = db.Model(models.Balance{}).Where("user_id = ?", userID).Select("balance").Updates(map[string]interface{}{"balance": newBalance})
	if r.Error != nil {
		return r.Error
	}
	return createOperation(db, userID, orderNumber, string(models.Accrual), amount)
}

func (s *storage) AccrualOrder(ctx context.Context, number string, newBalance float64, accrual float64) error {
	db := s.db.WithContext(ctx)
	db.Begin()
	defer db.Commit()
	order, err := updateOrder(db, number, string(models.OrderStatusProcessed), accrual)
	if err != nil {
		db.Rollback()
		return err
	}
	err = accrualBalance(db, order.UserID, number, newBalance, accrual)
	if err != nil {
		db.Rollback()
		return err
	}
	return err
}
func (s *storage) GetUserByOrderNumber(ctx context.Context, orderNumber string) (models.User, error) {
	db := s.db.WithContext(ctx)
	var order models.Order
	result := db.Where("number = ?", orderNumber).Preload("User").First(&order)
	return order.User, result.Error

}

func (s *storage) CreateOrder(ctx context.Context, userID uint, orderNumber string) (models.Order, error) {
	order := models.Order{UserID: uint(userID), Number: orderNumber, Status: models.OrderStatusNew}
	r := s.db.WithContext(ctx).Create(&order)
	return order, r.Error
}

func (s *storage) UpdateOrder(ctx context.Context, number string, status string, accrual float64) (models.Order, error) {
	db := s.db.WithContext(ctx)
	return updateOrder(db, number, status, accrual)
}

func updateOrder(db *gorm.DB, number string, status string, accrual float64) (models.Order, error) {
	if err := db.Model(models.Order{}).Where("number = ?", number).Select("status", "accrual").Updates(map[string]interface{}{"status": status, "accrual": accrual}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Order{}, internal_error.ErrOrderNotFound
		}
		return models.Order{}, err
	}
	return getOrder(db, number)
}

func (s *storage) GetProcessingOrders(ctx context.Context) ([]models.Order, error) {
	var orders []models.Order
	result := s.db.WithContext(ctx).Where("status = ? or status = ?", models.OrderStatusNew, models.OrderStatusProcessing).Find(&orders)
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return orders, internal_error.ErrOrderNotFound
	default:
		return orders, result.Error
	}
}

func getOrder(db *gorm.DB, orderNumber string) (models.Order, error) {
	var order models.Order
	result := db.Where("number = ?", orderNumber).First(&order)
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return order, internal_error.ErrOrderNotFound
	default:
		return order, result.Error
	}
}

func (s *storage) GetOrders(ctx context.Context, userID uint) ([]models.Order, error) {
	var orders []models.Order
	result := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&orders)
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return orders, internal_error.ErrOrderNotFound
	default:
		return orders, result.Error
	}
}

func (s *storage) GetOrder(ctx context.Context, orderNumber string) (models.Order, error) {
	db := s.db.WithContext(ctx)
	return getOrder(db, orderNumber)
}

func (s *storage) GetUser(ctx context.Context, login string) (models.User, error) {
	var user models.User
	result := s.db.WithContext(ctx).Where("name = ?", login).First(&user)
	switch {
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return user, internal_error.ErrUserNotFound
	default:
		return user, result.Error
	}
}

func (s *storage) CreateUser(ctx context.Context, login string, hash string) (models.User, error) {
	db := s.db.WithContext(ctx)
	db.Begin()
	defer db.Commit()
	var user models.User
	result := db.Where("name = ?", login).First(&user)
	switch {
	case result.Error == nil:
		return models.User{}, internal_error.ErrUserAlreadyExists
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		newUser := models.User{
			Name: login, Hash: hash,
		}
		r := db.Create(&newUser)
		if r.Error != nil {
			db.Rollback()
			return user, r.Error
		}
		balance := models.Balance{UserID: newUser.ID}
		r = db.Create(&balance)
		if r.Error != nil {
			db.Rollback()
		}
		user = newUser
		return user, r.Error
	default:
		return models.User{}, result.Error

	}
}

func (s *storage) GetBalance(ctx context.Context, userID uint) (models.Balance, error) {
	var balance models.Balance
	result := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&balance)
	if result.Error != nil {
		logger.Log.Info(result.Error.Error())
	}
	return balance, result.Error
}

func NewStorage(c *config.Config) storage {
	db, err := gorm.Open(postgres.Open(c.DBConnString), &gorm.Config{})
	if err != nil {
		logger.Log.Fatal(err.Error())
	}

	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Order{})
	db.AutoMigrate(&models.Balance{})
	db.AutoMigrate(&models.Operation{})
	return storage{db: db}
}
