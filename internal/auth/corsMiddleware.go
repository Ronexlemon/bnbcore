package auth

import (
	"net/http"
	"strings"
)

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowed := map[string]bool{
			"https://www.hostsasa.app":    true,
			"https://hostsasa.app":        true,
			"https://hostsasa.vercel.app": true,
			"http://localhost:3000":       true,
			"http://localhost:4000":       true,
			"http://localhost:5173":       true,
		}

		// Allow explicit origins
		isAllowed := allowed[origin]

		// Allow any *.hostsasa.app subdomain
		if strings.HasSuffix(origin, ".hostsasa.app") {
			isAllowed = true
		}

		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}