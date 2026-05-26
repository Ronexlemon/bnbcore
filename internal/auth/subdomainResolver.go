package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
)


func SubdomainResolver(service *tenant.Service, baseDomain string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            sub := extractSubdomain(r.Host, baseDomain)
			fmt.Printf("SubdomainResolver: host=%q base=%q extracted=%q\n", r.Host, baseDomain, sub)
            if sub == "" || sub == "www" {
				fmt.Println("SubdomainResolver: no subdomain — skipping tenant injection")
                next.ServeHTTP(w, r)
                return
            }
fmt.Println("Finding Subdomain")
            t, err := service.FindBySubdomain(r.Context(), sub)
            if err != nil || t == nil {
				fmt.Println("No shop Found")
                http.Error(w, "shop not found", http.StatusNotFound)
                return
            }
    fmt.Println("In the Resolver",t)
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