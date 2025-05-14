package storage

import "gorm.io/gorm"

type User struct {
	Name string
	Hash string
}

type Order struct {
	gorm.Model
	Number  int64
	Accrual int64
	Status  string
}

type Balance struct {
	User      User
	Balance   int64
	Withdrawn int64
}

type Operation struct {
	User  User
	Order Order
	Value int64
}
