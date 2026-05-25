package auth

import (
	"net/http"
	"strings"

	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
)


func SubdomainResolver(service *tenant.Service, baseDomain string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            sub := extractSubdomain(r.Host, baseDomain)
            if sub == "" || sub == "www" {
                next.ServeHTTP(w, r)
                return
            }

            t, err := service.FindBySubdomain(r.Context(), sub)
            if err != nil || t == nil {
                http.Error(w, "shop not found", http.StatusNotFound)
                return
            }

            // Use tenant.NewContext instead of raw context.WithValue
            ctx := tenant.NewContext(r.Context(), t)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func extractSubdomain(host, baseDomain string) string {
    host = strings.ToLower(strings.Split(host, ":")[0]) // strip port
    base := strings.ToLower(baseDomain)

    if !strings.HasSuffix(host, "."+base) {
        return ""
    }
    return strings.TrimSuffix(host, "."+base)
}