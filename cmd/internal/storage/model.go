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
	Accrual int64
	Status  OrderStatus
}

type Balance struct {
	gorm.Model
	UserID    uint
	User      User
	Balance   int64
	Withdrawn int64
}

type Operation struct {
	gorm.Model
	UserID        uint
	User          User
	OrderID       uint
	Order         Order
	OperationType string
	Value         int64
}
