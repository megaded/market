package router

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/megaded/market/cmd/internal/config"
	"github.com/megaded/market/cmd/internal/handler"
)

func CreateRouter(h handler.Handler, c config.Config) http.Handler {
	router := chi.NewRouter()
	router.Group(func(r chi.Router) {

		r.Use(middleware.Compress(5, "gzip"))
		r.Route("/api/user", func(rr chi.Router) {

			rr.Post("/orders", h.LoadOrder())
			rr.Get("/orders", h.Orders())
			rr.Get("/balance", h.Balance())
			rr.Post("/balance/withdraw", h.Withdraw())
			rr.Get("/withdrawals", h.Withdrawals())
		})
	})
	router.Group(func(r chi.Router) {
		r.Use(middleware.Compress(5, "gzip"))
		r.Post("/api/user/register", h.Register())
		r.Post("/api/user/login", h.Login())

	})

	return router
}
