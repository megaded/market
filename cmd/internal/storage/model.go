package storage

import "gorm.io/gorm"

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
	Number  uint
	Accrual int64
	Status  string
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
	UserID  uint
	User    User
	OrderID uint
	Order   Order
	Value   int64
}
