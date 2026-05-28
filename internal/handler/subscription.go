package handler

import (
	"encoding/json"
	"net/http"
	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/domain/subscription"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
)

type SubscriptionHandler struct {
	Server         *http.ServeMux
	Service        *subscription.Service
	JWTAuthManager *auth.JwtManager
}

func NewSubscriptionHandler(server *http.ServeMux, service *subscription.Service, m *auth.JwtManager) *SubscriptionHandler {
	h := &SubscriptionHandler{
		Server:         server,
		Service:        service,
		JWTAuthManager: m,
	}
	h.registerRoutes()
	return h
}

func (h *SubscriptionHandler) registerRoutes() {
	api := "/api/v1"

	// Public — show plans before login
	h.Server.HandleFunc("GET "+api+"/subscriptions/plans", h.GetPlans)

	// Protected
	h.Server.Handle("POST "+api+"/subscriptions",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(h.Subscribe)))

	h.Server.Handle("GET "+api+"/subscriptions/me",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(h.GetMySubscription)))

	h.Server.Handle("PUT "+api+"/subscriptions/upgrade",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(h.Upgrade)))

	h.Server.Handle("PATCH "+api+"/subscriptions/cancel",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(h.Cancel)))
}


func (h *SubscriptionHandler) GetPlans(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{
		"currency": "KES",
		"plans":    h.Service.GetPlans(),
	})
}


func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req subscription.CreateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tenant := tenant.FromContext(r.Context())
	if tenant.ID ==nil{
		 http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return

	}

	tenantID :=*tenant.ID

	result, err := h.Service.Subscribe(r.Context(), tenantID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{
		"message": "subscribed successfully",
		"data":    result,
	})
}

func (h *SubscriptionHandler) GetMySubscription(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tenant := tenant.FromContext(r.Context())
	if tenant.ID ==nil{
		 http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return

	}

	tenantID :=*tenant.ID


	result, err := h.Service.GetMySubscription(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]any{"data": result})
}


func (h *SubscriptionHandler) Upgrade(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req subscription.UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tenant := tenant.FromContext(r.Context())
	if tenant.ID ==nil{
		 http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return

	}

	tenantID :=*tenant.ID


	result, err := h.Service.Upgrade(r.Context(), tenantID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]any{
		"message": "subscription upgraded successfully",
		"data":    result,
	})
}


func (h *SubscriptionHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	tenant := tenant.FromContext(r.Context())
	if tenant.ID ==nil{
		 http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return

	}

	tenantID :=*tenant.ID


	if err := h.Service.Cancel(r.Context(), tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]any{
		"message": "subscription canceled successfully",
	})
}