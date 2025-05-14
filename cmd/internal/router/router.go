package router

import (
	"net/http"

	"github.com/go-chi/chi"
)

func CreateRouter() http.Handler {
	router := chi.NewRouter()
	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", nil)
		r.Post("/login", nil)
		r.Post("/orders", nil)
		r.Get("/orders", nil)
		r.Get("/balance", nil)
		r.Post("/balance/withdraw", nil)
		r.Get("/withdrawals", nil)
	})
	return router
}
