package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"


)

// contextKey is an unexported type for context keys in this package.
type contextKey string

const (
	// ClaimsKey is the key used to store *Claims in the request context.
	ClaimsKey contextKey = "jwt_claims"
)

// ClaimsFromContext retrieves *Claims stored by the middleware.
// Returns nil if the context carries no claims.
func ClaimsFromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(ClaimsKey).(*Claims)
	return c
}


// Authenticate returns a middleware that validates the Bearer token in the
// Authorization header and stores the claims in the request context.
func (m *JwtManager) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := m.extractClaims(r)
		if err != nil {
			http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns a middleware that allows access only when the authenticated
// user has one of the given roles. Must be chained after Authenticate.
func (m *JwtManager) RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if _, ok := allowed[claims.Role]; !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}


// extractClaims pulls the Bearer token from the Authorization header and
// validates it, returning the embedded Claims.
func (m *JwtManager) extractClaims(r *http.Request) (*Claims, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, errorf("missing Authorization header")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return nil, errorf("authorization header must be 'Bearer <token>'")
	}

	return m.ValidateToken(parts[1])
}

func errorf(msg string) error {
	return errors.New(msg)
}