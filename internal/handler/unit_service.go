package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/auth"
	rs "github.com/ronexlemon/bnbcore/internal/domain/services"
	"github.com/ronexlemon/bnbcore/internal/domain/subscription"
	"github.com/ronexlemon/bnbcore/internal/eventstream"
	"github.com/ronexlemon/bnbcore/internal/middleware"
)

type RoomServiceHandler struct {
    Server         *http.ServeMux
    Service        *rs.Service
    JWTAuthManager *auth.JwtManager
	SubRepo        subscription.Repository
	 Stream         *eventstream.KafkaClient
}

func NewRoomServiceHandler(server *http.ServeMux, service *rs.Service, m *auth.JwtManager,sub  subscription.Repository,stream  *eventstream.KafkaClient ) *RoomServiceHandler {
    h := &RoomServiceHandler{
        Server:         server,
        Service:        service,
        JWTAuthManager: m,
		SubRepo: sub,
		Stream: stream,
    }
    h.registerRoutes()
    return h
}

func (h *RoomServiceHandler) registerRoutes() {
    api := "/api/v1"

	protected := func(hf http.HandlerFunc) http.Handler {
		return h.JWTAuthManager.Authenticate(
			middleware.RequireActiveSubscription(h.SubRepo)(hf),
		)
	}
    h.Server.Handle("POST "+api+"/units/{unit_id}/unit-services",protected(h.Create))

    h.Server.Handle("GET "+api+"/units/{unit_id}/unit-services",
        h.JWTAuthManager.Authenticate(http.HandlerFunc(h.GetByUnit)))

    h.Server.Handle("GET "+api+"/unit-services/{id}",
        h.JWTAuthManager.Authenticate(http.HandlerFunc(h.GetByID)))

    h.Server.Handle("PUT "+api+"/unit-services/{id}",
        h.JWTAuthManager.Authenticate(http.HandlerFunc(h.Update)))

    h.Server.Handle("DELETE "+api+"/unit-services/{id}",protected(h.Delete))
       
}

func (h *RoomServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
    claims := auth.ClaimsFromContext(r.Context())
    if claims == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    unitID, err := uuid.Parse(r.PathValue("unit_id"))
    if err != nil {
        http.Error(w, "invalid unit id", http.StatusBadRequest)
        return
    }

  if claims.TenantID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *claims.TenantID

    var req rs.CreateUnitServiceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    result, err := h.Service.Create(r.Context(), unitID, tenantID, req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.WriteHeader(http.StatusCreated)
    writeJSON(w, map[string]any{
        "message": "room service created successfully",
        "data":    result,
    })
}

func (h *RoomServiceHandler) GetByUnit(w http.ResponseWriter, r *http.Request) {
    claims := auth.ClaimsFromContext(r.Context())
    if claims == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    unitID, err := uuid.Parse(r.PathValue("unit_id"))
    if err != nil {
        http.Error(w, "invalid unit id", http.StatusBadRequest)
        return
    }

if claims.TenantID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *claims.TenantID

    services, err := h.Service.GetByUnit(r.Context(), unitID, tenantID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]any{"data": services})
}

func (h *RoomServiceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    claims := auth.ClaimsFromContext(r.Context())
    if claims == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    id, err := uuid.Parse(r.PathValue("id"))
    if err != nil {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }

    if claims.TenantID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *claims.TenantID

    result, err := h.Service.GetByID(r.Context(), id, tenantID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    writeJSON(w, map[string]any{"data": result})
}

func (h *RoomServiceHandler) Update(w http.ResponseWriter, r *http.Request) {
    claims := auth.ClaimsFromContext(r.Context())
    if claims == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    id, err := uuid.Parse(r.PathValue("id"))
    if err != nil {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }

   if claims.TenantID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *claims.TenantID

    var req rs.UpdateUnitServiceRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request body", http.StatusBadRequest)
        return
    }

    result, err := h.Service.Update(r.Context(), id, tenantID, req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    writeJSON(w, map[string]any{
        "message": "room service updated successfully",
        "data":    result,
    })
}

func (h *RoomServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
    claims := auth.ClaimsFromContext(r.Context())
    if claims == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    id, err := uuid.Parse(r.PathValue("id"))
    if err != nil {
        http.Error(w, "invalid id", http.StatusBadRequest)
        return
    }

    if claims.TenantID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *claims.TenantID

    if err := h.Service.Delete(r.Context(), id, tenantID); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    w.WriteHeader(http.StatusNoContent)
}