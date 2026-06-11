package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/domain/subscription"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/eventstream"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/upload"
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
    Media         *upload.MediaService
}

func NewTenantHandler(server *http.ServeMux, service *tenant.Service, m *auth.JwtManager,sub subscription.Repository,stream *eventstream.KafkaClient,media *upload.MediaService) *TenantHandler {
	h := &TenantHandler{
		Server:         server,
		Service:        service,
		JWTAuthManager: m,
		SubRepo: sub,
		Stream: stream,
		Media: media,
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
	if claims.UserID == nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid tenant id", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB total — allows up to 2 images
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "file size too large or malformed form data", http.StatusBadRequest)
		return
	}

	req := tenant.UpdateTenantRequest{}

	if v := r.FormValue("shop_description"); v != "" {
		req.ShopDescription = &v
	}
	if v := r.FormValue("subdomain"); v != "" {
		req.Subdomain = &v
	}
	if v := r.FormValue("name"); v != "" {
		req.ShopName = &v
	}
	if v := r.FormValue("phone_number"); v != "" {
		req.PhoneNumber = &v
	}
	if v := r.FormValue("long_description"); v != "" {
		req.LongDescription = &v
	}
	

	// ── Banner upload (optional)
	if file, header, err := r.FormFile("banner"); err == nil {
		defer file.Close()
		cacheKey := fmt.Sprintf("tenant:%s:banner:%s", id.String(), header.Filename)
		url, uploadErr := h.Media.UploadAndCacheStream(r.Context(), file, cacheKey, "hostsasa")
		if uploadErr != nil {
			http.Error(w, fmt.Sprintf("failed to upload banner: %v", uploadErr), http.StatusInternalServerError)
			return
		}
		req.Banner = &url
	}

	// ── Profile photo upload (optional)
	if file, header, err := r.FormFile("profile_photo"); err == nil {
		defer file.Close()
		cacheKey := fmt.Sprintf("tenant:%s:profile_photo:%s", id.String(), header.Filename)
		url, uploadErr := h.Media.UploadAndCacheStream(r.Context(), file, cacheKey, "hostsasa")
		if uploadErr != nil {
			http.Error(w, fmt.Sprintf("failed to upload profile photo: %v", uploadErr), http.StatusInternalServerError)
			return
		}
		req.ProfilePhoto = &url
	}

	result, err := h.Service.UpdateTenant(r.Context(), id, req)
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
	if req.ShopName != nil {
		changes["name"] = *req.ShopName
	}
	if req.PhoneNumber != nil {
		changes["phone_number"] = *req.PhoneNumber
	}
	if req.LongDescription != nil {
		changes["long_description"] = *req.LongDescription
	}

	if req.Banner != nil {
		changes["banner"] = *req.Banner
	}
	if req.ProfilePhoto != nil {
		changes["profile_photo"] = *req.ProfilePhoto
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