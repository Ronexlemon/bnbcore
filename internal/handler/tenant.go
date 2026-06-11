package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/domain/subscription"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/eventstream"
	"github.com/ronexlemon/bnbcore/internal/metrics"
)

type RegisterTenantRequest struct {
	ShopDescription  string `json:"shop_description"`
	Subdomain string `json:"subdomain"`
	ShopName string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	LongDescription string `json:"long_description"`
}



type TenantHandler struct {
	Server         *http.ServeMux
	Service        *tenant.Service
	JWTAuthManager *auth.JwtManager
	SubRepo        subscription.Repository
		 Stream         *eventstream.KafkaClient
}

func NewTenantHandler(server *http.ServeMux, service *tenant.Service, m *auth.JwtManager,sub subscription.Repository,stream *eventstream.KafkaClient) *TenantHandler {
	h := &TenantHandler{
		Server:         server,
		Service:        service,
		JWTAuthManager: m,
		SubRepo: sub,
		Stream: stream,
	}
	h.registerHandler()
	return h
}

func (h *TenantHandler) registerHandler() {
	api := "/api/v1"

	// protected := func(hf http.HandlerFunc) http.Handler {
	// 	return h.JWTAuthManager.Authenticate(
	// 		middleware.RequireActiveSubscription(h.SubRepo)(hf),
	// 	)
	// }
	h.Server.Handle("POST "+api+"/tenant",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.CreateTenant))))

	h.Server.Handle("GET "+api+"/tenant",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.GetTenant))))

	h.Server.Handle("PUT "+api+"/tenant/{id}",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.UpdateTenant))))

	h.Server.Handle("DELETE "+api+"/tenant/{id}",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.DeleteTenant))))

	h.Server.Handle("GET "+api+"/tenant/{id}",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.GetTenantByID))))

	h.Server.Handle("GET "+api+"/tenant/subdomain-availability",
		http.HandlerFunc(metrics.MetricsMiddleware(h.CheckSubdomainAvailability)))
}

func (h *TenantHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req RegisterTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ShopDescription == "" || req.Subdomain == "" {
		http.Error(w, "shop_name and subdomain are required", http.StatusBadRequest)
		return
	}

	if claims.UserID == nil {
        http.Error(w,"invalid user id" ,http.StatusBadRequest)
        return
    }
userID := *claims.UserID


	 tenant_details,err := h.Service.CreateTenant(r.Context(),req.ShopName ,req.ShopDescription, req.Subdomain,req.PhoneNumber, userID,req.LongDescription)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
   _ = h.Stream.Publish(r.Context(), eventstream.TopicTenantCreated, tenant_details.ID.String(),
    eventstream.TenantCreatedEvent{
        BaseEvent: eventstream.BaseEvent{
            TenantID:  *tenant_details.Tenant.ID,
            UserID:    userID,
            UserEmail: claims.Email,
            ShopName:  req.ShopName,
            OccuredAt: time.Now(),
        },
        Subdomain: req.Subdomain,
    },
)

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{
		"message": "tenant created successfully",
		"data":tenant_details,
	})
}

func (h *TenantHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.UserID == nil {
        http.Error(w,"invalid user id" ,http.StatusBadRequest)
        return
    }
userID := *claims.UserID

	result, err := h.Service.GetTenantByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]any{
		"data": result,
	})
}

func (h *TenantHandler) GetTenantByID(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid tenant id", http.StatusBadRequest)
		return
	}

	result, err := h.Service.GetTenantByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]any{
		"data": result,
	})
}


func (h *TenantHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid tenant id", http.StatusBadRequest)
		return
	}

	var req tenant.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var status *tenant.TenantStatus
	if req.Status != nil {
		s := tenant.TenantStatus(*req.Status)
		status = &s
	}

	result, err := h.Service.UpdateTenant(r.Context(), id, tenant.UpdateTenantRequest{
		ShopDescription:  req.ShopDescription,
		Subdomain: req.Subdomain,
		Status:    status,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	changes := map[string]any{}
	if req.ShopDescription != nil {
		changes["shop_description"] = *req.ShopDescription
	}
	if req.Subdomain != nil {
		changes["subdomain"] = *req.Subdomain
	}
	if req.Status != nil {
		changes["status"] = *req.Status
	}

	_ = h.Stream.Publish(r.Context(), eventstream.TopicTenantUpdated, result.ID.String(),
		eventstream.TenantUpdatedEvent{
			BaseEvent: eventstream.BaseEvent{
				TenantID:  *result.ID,
				UserID:    *claims.UserID,
				UserEmail: claims.Email,
				ShopName:  *result.ShopName,
				OccuredAt: time.Now(),
			},
			Changes: changes,
			
		},
	)

	writeJSON(w, map[string]any{
		"message": "tenant updated successfully",
		"data":    result,
	})
}
func (h *TenantHandler) CheckSubdomainAvailability(w http.ResponseWriter, r *http.Request) {
	subdomain := strings.TrimSpace(r.URL.Query().Get("subdomain"))
	if subdomain == "" {
		http.Error(w, "subdomain query parameter is required", http.StatusBadRequest)
		return
	}

	exists, err := h.Service.SubdomainExists(r.Context(), subdomain)
	if err != nil {
		http.Error(w, "failed to check subdomain availability", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"subdomain": subdomain,
		"available": !exists,
	})
}
func (h *TenantHandler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid tenant id", http.StatusBadRequest)
		return
	}

	if err := h.Service.DeleteTenant(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = h.Stream.Publish(r.Context(), eventstream.TopicTenantDeleted, id.String(),
    eventstream.TenantDeletedEvent{
        BaseEvent: eventstream.BaseEvent{
            UserID:    *claims.UserID,
            UserEmail: claims.Email,
            OccuredAt: time.Now(),
        },
    },
)

	w.WriteHeader(http.StatusNoContent)
}