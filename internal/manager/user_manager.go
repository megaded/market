package manager

import (
	"context"

	internalerror "github.com/megaded/market/internal/error"
	"github.com/megaded/market/internal/identity"
	"github.com/megaded/market/internal/storage"
	"github.com/megaded/market/internal/storage/models"
)

type UserStorager interface {
	CreateUser(ctx context.Context, login string, password string) (models.User, error)
}

type UserManager struct {
	storage  UserStorager
	identity identity.IdentityProvider
}

func CreateUserManager(s storage.Storager) UserManager {
	return UserManager{storage: s}
}

func (u *UserManager) CreateUser(ctx context.Context, login string, password string) (models.User, error) {
	if login == "" || password == "" {
		return models.User{}, internalerror.ErrEmptyLoginOrPassword
	}
	hash := u.identity.HashPassword(password)
	return u.storage.CreateUser(ctx, login, hash)
}
