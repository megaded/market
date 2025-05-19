package handler

import (
	"net/http"

	"github.com/megaded/market/cmd/internal/storage"
)

type Handler struct {
	Storage storage.Storager
}

func CreateHandlers(s storage.Storager) Handler {
	return Handler{Storage: s}
}

func (h *Handler) Register() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

func (h *Handler) Login() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

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
