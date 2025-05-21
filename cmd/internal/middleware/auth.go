package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/megaded/market/cmd/internal/identity"
)

func AuthMiddleWare(id identity.IdentityProvider) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(token, prefix) {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			token = strings.TrimPrefix(token, prefix)

			userID, err := id.ParseToken(token)
			if err != nil {
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
