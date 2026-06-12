package middleware

import (
	"net/http"

	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/domain/subscription"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
)

func RequireActiveSubscription(repo subscription.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant := tenant.FromContext(r.Context())
			if tenant == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

           
			if tenant.UserID == nil {
				http.Error(w, "no tenant associated with this account", http.StatusForbidden)
				return
			}

			userID :=*tenant.UserID

			active, err := repo.IsActive(r.Context(), userID)
			if err != nil {
				http.Error(w, "failed to verify subscription", http.StatusInternalServerError)
				return
			}
			if !active {
				http.Error(w, "subscription expired or inactive — please renew your plan", http.StatusPaymentRequired) // 402
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func RequireActiveOnDirectSubscription(repo subscription.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := auth.ClaimsFromContext(r.Context())
			if claims == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			active, err := repo.IsActive(r.Context(), *claims.UserID)
			if err != nil {
				http.Error(w, "failed to verify subscription", http.StatusInternalServerError)
				return
			}
			if !active {
				http.Error(w, "subscription expired or inactive — please renew your plan", http.StatusPaymentRequired) // 402
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}