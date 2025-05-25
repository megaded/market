package internal_error

import "errors"

var (
	ErrOrderNotFound                    = errors.New("order not found")
	ErrInvalidOrderNumber               = errors.New("invalid order number")
	ErrOrderAlreadyExists               = errors.New("order already exists")
	ErrOrderAlreadyExistsForAnotherUser = errors.New("order already exists for another user")
	ErrInvalidWithdrawSum               = errors.New("invalid withdraw sum")
	ErrWithdrawalNotFound               = errors.New("withdrawal not found")
	ErrEmptyLoginOrPassword             = errors.New("login or password is empty")
	ErrUserAlreadyExists                = errors.New("user already exists")
	ErrUserNotFound                     = errors.New("user not found")
	ErrInvalidPassword                  = errors.New("invalid password")
)
