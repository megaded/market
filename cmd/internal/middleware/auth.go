package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/megaded/market/cmd/internal/identity"
	"github.com/megaded/market/cmd/internal/logger"
)

func AuthMiddleWare(id identity.IdentityProvider) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			logger.Log.Info("auth")
			token := r.Header.Get("Authorization")
			logger.Log.Info(token)
			if token == "" {
				logger.Log.Info("token")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(token, prefix) {
				logger.Log.Info("Bearer")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			token = strings.TrimPrefix(token, prefix)

			userID, err := id.ParseToken(token)
			if err != nil {
				logger.Log.Info(err.Error())
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, identity.UserID, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}
