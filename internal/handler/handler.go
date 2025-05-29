package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/megaded/market/internal/dto"
	internal_error "github.com/megaded/market/internal/error"
	"github.com/megaded/market/internal/identity"
	"github.com/megaded/market/internal/logger"
	"github.com/megaded/market/internal/manager"
	"github.com/megaded/market/internal/storage/models"
)

type Storager interface {
	GetUser(ctx context.Context, login string) (models.User, error)
	GetOrders(ctx context.Context, userID uint) ([]models.Order, error)
	GetBalance(ctx context.Context, userID uint) (models.Balance, error)
	GetOperations(ctx context.Context, userID uint, operationType string) ([]models.Operation, error)
}

type Handler struct {
	Storage      Storager
	Identity     identity.IdentityProvider
	OrderManager manager.OrderManager
	UserManager  manager.UserManager
}

func getUserID(r *http.Request) (int, error) {
	userID, ok := r.Context().Value(identity.UserID).(int)
	if !ok {
		return 0, errors.New("user ID not found in context")
	}
	return userID, nil
}

func CreateHandlers(s Storager, m manager.OrderManager, u manager.UserManager) Handler {
	return Handler{Storage: s, OrderManager: m, UserManager: u}
}

func (h *Handler) Register() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var user dto.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			handleError(w, http.StatusBadRequest, err)
			return
		}
		newUser, err := h.UserManager.CreateUser(r.Context(), user.Login, user.Password)
		if err != nil {
			switch {
			case errors.Is(err, internal_error.ErrUserAlreadyExists):
				handleError(w, http.StatusConflict, err)
				return
			default:
				handleError(w, http.StatusInternalServerError, err)
				return
			}
		}
		token, err := h.Identity.GenerateToken(int(newUser.ID))
		if err != nil {
			handleError(w, http.StatusInternalServerError, err)
			return
		}

		w.Header().Set("Authorization", fmt.Sprintf("Bearer %s", token))
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) Login() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var user dto.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			handleError(w, http.StatusBadRequest, err)
			return
		}
		if user.Login == "" || user.Password == "" {
			handleError(w, http.StatusUnauthorized, err)
			return
		}
		userInfo, err := h.Storage.GetUser(r.Context(), user.Login)
		switch {
		case err == nil:
			valResult := h.Identity.VerifyPassword(userInfo.Hash, user.Password)
			if valResult {
				token, err := h.Identity.GenerateToken(int(userInfo.ID))
				if err != nil {
					handleError(w, http.StatusInternalServerError, err)
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
			handleError(w, http.StatusUnauthorized, err)
			return
		default:
			handleError(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (h *Handler) LoadOrder() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserID(r)
		if err != nil {
			handleError(w, http.StatusUnauthorized, err)
			return
		}

		if r.Header.Get("Content-Type") != "text/plain" {
			handleError(w, http.StatusBadRequest, internal_error.ErrInvalidContentType)
			return
		}

		orderNumber, err := io.ReadAll(r.Body)
		if err != nil {
			handleError(w, http.StatusBadRequest, err)
			return
		}

		if len(orderNumber) == 0 {
			handleError(w, http.StatusBadRequest, internal_error.ErrOrderNotFound)
			return
		}

		if err = h.OrderManager.AddOrder(r.Context(), uint(userID), string(orderNumber)); err != nil {
			switch {
			case errors.Is(err, internal_error.ErrInvalidOrderNumber):
				handleError(w, http.StatusUnprocessableEntity, err)
				return
			case errors.Is(err, internal_error.ErrOrderAlreadyExists):
				w.WriteHeader(http.StatusOK)
				return
			case errors.Is(err, internal_error.ErrOrderAlreadyExistsForAnotherUser):
				handleError(w, http.StatusConflict, err)
				return
			default:
				handleError(w, http.StatusInternalServerError, err)
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
			handleError(w, http.StatusUnauthorized, err)
			return
		}
		orders, err := h.Storage.GetOrders(r.Context(), uint(userID))
		if err != nil {
			switch {
			case errors.Is(err, internal_error.ErrOrderNotFound):
				handleError(w, http.StatusNoContent, err)
				return
			default:
				handleError(w, http.StatusInternalServerError, err)
				return
			}
		}
		result := make([]dto.Order, 0, len(orders))
		for _, op := range orders {
			result = append(result, dto.Order{Number: op.Number, Status: string(op.Status), UploadedAt: op.CreatedAt, Accrual: op.Accrual})
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(result); err != nil {
			handleError(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (h *Handler) Balance() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID, err := getUserID(r)
		if err != nil {
			handleError(w, http.StatusUnauthorized, err)
			return
		}
		balance, err := h.Storage.GetBalance(r.Context(), uint(userID))
		if err != nil {
			handleError(w, http.StatusInternalServerError, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(dto.Balance{Current: balance.Balance, Withdrawn: balance.Withdrawn}); err != nil {
			handleError(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (h *Handler) Withdraw() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserID(r)
		if err != nil {
			handleError(w, http.StatusUnauthorized, err)
			return
		}
		var withDraw dto.Withdraw
		err = json.NewDecoder(r.Body).Decode(&withDraw)
		if err != nil {
			handleError(w, http.StatusBadRequest, err)
			return
		}
		err = h.OrderManager.WithdrawOrder(r.Context(), uint(userID), withDraw.Order, withDraw.Sum)
		if err != nil {
			switch {
			case errors.Is(err, internal_error.ErrInvalidWithdrawSum):
				handleError(w, http.StatusPaymentRequired, err)
				return
			case errors.Is(err, internal_error.ErrInvalidOrderNumber):
				handleError(w, http.StatusUnprocessableEntity, err)
				return
			default:
				handleError(w, http.StatusInternalServerError, err)
			}

		}
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handler) Withdrawals() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserID(r)
		if err != nil {
			handleError(w, http.StatusUnauthorized, err)
			return
		}
		operations, err := h.Storage.GetOperations(r.Context(), uint(userID), string(models.Withdraw))
		if err != nil {
			switch {
			case errors.Is(err, internal_error.ErrWithdrawalNotFound):
				handleError(w, http.StatusNoContent, err)
				return
			default:
				handleError(w, http.StatusInternalServerError, err)
				return
			}
		}
		result := make([]dto.Withdraw, 0, len(operations))
		for _, op := range operations {
			result = append(result, dto.Withdraw{Order: op.OrderNumber, Sum: float64(op.Value), ProcessedAt: op.CreatedAt})
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(result); err != nil {
			handleError(w, http.StatusInternalServerError, err)
			return
		}
	}
}

func handleError(w http.ResponseWriter, errorCode int, err error) {
	w.WriteHeader(errorCode)
	w.Write([]byte(err.Error()))
	logger.Log.Error(err.Error())
}
