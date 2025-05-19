package storage

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/megaded/market/cmd/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Storager interface {
	GetOrders(userId int64) ([]Order, error)
	GetBalance(userId int64) (Balance, error)
	CreateUser(login string, hash string) error
}

type storage struct {
	db  *gorm.DB
	key string
}

func (s *storage) GetOrders(userId int64) ([]Order, error) {
	return nil, nil
}
func (s *storage) CreateUser(login string, password string) error {
	r := s.db.Create(&User{Name: login, Hash: hash(password, s.key)})
	return r.Error
}

func (s *storage) GetBalance(userId int64) (Balance, error) {
	return Balance{}, nil
}

func NewStorage(c *config.Config) Storager {
	db, err := gorm.Open(postgres.Open(c.DbConnString), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&User{})
	db.AutoMigrate(&Order{})
	db.AutoMigrate(&Balance{})
	db.AutoMigrate(&Operation{})
	return &storage{}
}

func hash(password string, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}
