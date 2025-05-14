package dto

import "gorm.io/gorm"

type UserDto struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

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
