package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/megaded/market/cmd/internal/dto"
	internal_error "github.com/megaded/market/cmd/internal/error"
	"github.com/megaded/market/cmd/internal/identity"
	"github.com/megaded/market/cmd/internal/logger"
	"github.com/megaded/market/cmd/internal/manager"
	"github.com/megaded/market/cmd/internal/storage"
	"go.uber.org/zap"
)

type Handler struct {
	Storage      storage.Storager
	Identity     identity.IdentityProvider
	OrderManager manager.OrderManager
}

func getUserID(r *http.Request) (uint, error) {
	userID, ok := r.Context().Value(identity.UserID).(uint)
	if !ok {
		return 0, errors.New("user ID not found in context")
	}
	return userID, nil
}

func CreateHandlers(s storage.Storager, m manager.OrderManager) Handler {
	return Handler{Storage: s, OrderManager: m}
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
		newUser, err := h.Storage.CreateUser(user.Login, user.Password)
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
		if user.Login == "" || user.Password == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		userInfo, err := h.Storage.GetUser(user.Login)
		switch {
		case err == nil:
			valResult := h.Identity.VerifyPassword(userInfo.Hash, user.Password)
			if valResult {
				token, err := h.Identity.GenerateToken(int(userInfo.ID))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				logger.Log.Info(fmt.Sprintf("User %s Authorization", user.Login))
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
		logger.Log.Info("LoadOrder")
		userID, err := getUserID(r)
		if err != nil {
			logger.Log.Info(err.Error())
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Header.Get("Content-Type") != "text/plain" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		orderNumber, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(orderNumber) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if err = h.OrderManager.AddOrder(int64(userID), string(orderNumber)); err != nil {
			switch {
			case errors.Is(err, internal_error.ErrInvalidOrderNumber):
				logger.Log.Info(string(orderNumber))
				logger.Log.Info(err.Error())
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			case errors.Is(err, internal_error.ErrOrderAlreadyExists):
				logger.Log.Info(string(orderNumber))
				logger.Log.Info(err.Error())
				w.WriteHeader(http.StatusOK)
				return
			case errors.Is(err, internal_error.ErrOrderAlreadyExistsForAnotherUser):
				logger.Log.Info(string(orderNumber))
				logger.Log.Info(err.Error())
				w.WriteHeader(http.StatusConflict)
				return
			default:
				w.WriteHeader(http.StatusInternalServerError)
				logger.Log.Info(string(orderNumber))
				logger.Log.Info("failed to add order", zap.Error(err))
				return
			}
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *Handler) Orders() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserID(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		orders, err := h.Storage.GetOrders(int64(userID))
		if err != nil {
			switch {
			case errors.Is(err, internal_error.ErrOrderNotFound):
				w.WriteHeader(http.StatusNoContent)
				logger.Log.Info("orders not found", zap.Error(err))
				return
			default:
				w.WriteHeader(http.StatusInternalServerError)
				logger.Log.Info("internal error", zap.Error(err))
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(orders); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Log.Info("failed to encode orders", zap.Error(err))
			return
		}
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
