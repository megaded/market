package storage

import "gorm.io/gorm"

type OrderStatus string
type OperationType string

const (
	Accrual  OperationType = "Accrual"
	Withdraw OperationType = "Withdraw"
)

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type User struct {
	gorm.Model
	Name      string
	Hash      string
	Operation []Operation
}

type Order struct {
	gorm.Model
	UserID uint
	User
	Number  string
	Accrual float64
	Status  OrderStatus
}

type Balance struct {
	gorm.Model
	UserID    uint
	User      User
	Balance   float64
	Withdrawn float64
}

type Operation struct {
	gorm.Model
	UserID        uint
	User          User
	OrderID       uint `gorm:"default:null"`
	Order         Order
	OperationType string
	Value         float64
}
