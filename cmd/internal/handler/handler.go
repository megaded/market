package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/megaded/market/cmd/internal/dto"
	internal_error "github.com/megaded/market/cmd/internal/error"
	"github.com/megaded/market/cmd/internal/identity"
	"github.com/megaded/market/cmd/internal/logger"
	"github.com/megaded/market/cmd/internal/storage"
)

type Handler struct {
	Storage  storage.Storager
	Identity identity.IdentityProvider
}

func CreateHandlers(s storage.Storager) Handler {
	return Handler{Storage: s}
}

func (h *Handler) Register() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var user dto.UserDto
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Log.Info(err.Error())
			w.Write([]byte(err.Error()))
			return
		}
		newUser, err := h.Storage.CreateUser(user.Name, user.Password)
		if err != nil {
			switch {
			case errors.Is(err, internal_error.ErrUserAlreadyExists):
				w.WriteHeader(http.StatusConflict)
				logger.Log.Info(err.Error())
				w.Write([]byte(err.Error()))
			default:
				w.WriteHeader(http.StatusInternalServerError)
				logger.Log.Info(err.Error())
				w.Write([]byte(err.Error()))
			}
		}
		token, err := h.Identity.GenerateToken(int(newUser.ID))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Info(err.Error())
			w.Write([]byte(err.Error()))
		}

		w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) Login() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var user dto.UserDto
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Log.Info(err.Error())
			w.Write([]byte(err.Error()))
			return
		}
		if user.Name == "" || user.Password == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userInfo, err := h.Storage.GetUser(user.Name)
		switch {
		case err == nil:
			valResult := h.Identity.VerifyPassword(userInfo.Hash, user.Password)
			if valResult {
				token, err := h.Identity.GenerateToken(int(userInfo.ID))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				logger.Log.Info(fmt.Sprintf("User %s Authorization", user.Name))
				w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		case errors.Is(err, internal_error.ErrUserNotFound):
			w.WriteHeader(http.StatusUnauthorized)
			logger.Log.Info(err.Error())
			w.Write([]byte(err.Error()))
			return
		default:
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Info(err.Error())
			w.Write([]byte(err.Error()))
			return
		}
	}
}

func (h *Handler) LoadOrder() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *Handler) Orders() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *Handler) Balance() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *Handler) Withdraw() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *Handler) Withdrawals() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
