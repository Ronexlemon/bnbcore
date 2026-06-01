package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/domain/subscription"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/domain/unit"
	"github.com/ronexlemon/bnbcore/internal/eventstream"
	"github.com/ronexlemon/bnbcore/internal/middleware"
)

type UnitHandler struct {
	Server         *http.ServeMux
	Service        *unit.UnitService
	JWTAuthManager *auth.JwtManager
	SubRepo        subscription.Repository
	Stream         *eventstream.KafkaClient
}

func NewUnitHandler(server *http.ServeMux, service *unit.UnitService, m *auth.JwtManager,sub subscription.Repository,stream  *eventstream.KafkaClient) *UnitHandler {
	h := &UnitHandler{
		Server:         server,
		Service:        service,
		JWTAuthManager: m,
		SubRepo: sub,
		Stream: stream,
	}
	h.registerRoutes()
	return h
}

func (h *UnitHandler) registerRoutes() {
	api := "/api/v1"

	protected := func(hf http.HandlerFunc) http.Handler {
		return h.JWTAuthManager.Authenticate(
			middleware.RequireActiveSubscription(h.SubRepo)(hf),
		)
	}
	
	h.Server.HandleFunc("GET "+api+"/units", h.GetAllUnits)
	h.Server.HandleFunc("GET "+api+"/units/{id}", h.GetUnit)

	h.Server.Handle("POST "+api+"/units",protected(h.CreateUnit))


	h.Server.Handle("PUT "+api+"/units/{id}",protected(h.UpdateUnit))

	h.Server.Handle("DELETE "+api+"/units/{id}",protected((h.DeleteUnit)))

	h.Server.Handle("POST "+api+"/units/{id}/images",protected(h.AddImage))

	h.Server.Handle("DELETE "+api+"/units/{id}/images/{image_id}",
	 h.JWTAuthManager.Authenticate(http.HandlerFunc(h.RemoveImage)))

}

func (h *UnitHandler) CreateUnit(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req unit.CreateUnitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
   
	tenant:=tenant.FromContext(r.Context())

	if tenant.ID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *tenant.ID

	result, err := h.Service.CreateUnit(r.Context(), tenantID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_ = h.Stream.Publish(r.Context(), eventstream.TopicUnitCreated, result.ID.String(),
    eventstream.UnitCreatedEvent{
        BaseEvent: eventstream.BaseEvent{
            TenantID:  *tenant.ID,
            UserID:    *claims.UserID,
            UserEmail: claims.Email,
            OccuredAt: time.Now(),
        },
        Title: result.Title,
    },
)

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{
		"message": "unit created successfully",
		"data":    result,
	})
}

func (h *UnitHandler) GetAllUnits(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	
	if t == nil {
		http.Error(w, "tenant not found", http.StatusBadRequest)
		return
	}

	if t.ID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *t.ID
	 
	units, err := h.Service.GetAllUnits(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"data": units,
	})
}

func (h *UnitHandler) GetUnit(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		http.Error(w, "tenant not found", http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid unit id", http.StatusBadRequest)
		return
	}
	if t.ID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *t.ID

	result, err := h.Service.GetUnit(r.Context(), id, tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]any{
		"data": result,
	})
}

func (h *UnitHandler) UpdateUnit(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid unit id", http.StatusBadRequest)
		return
	}

	var req unit.UpdateUnitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tenant:=tenant.FromContext(r.Context())

	if tenant.ID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *tenant.ID

	result, err := h.Service.UpdateUnit(r.Context(), id, tenantID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]any{
		"message": "unit updated successfully",
		"data":    result,
	})
}

func (h *UnitHandler) DeleteUnit(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid unit id", http.StatusBadRequest)
		return
	}

	if claims.TenantID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *claims.TenantID

	if err := h.Service.DeleteUnit(r.Context(), id, tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UnitHandler) AddImage(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	unitID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid unit id", http.StatusBadRequest)
		return
	}

	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	image := &unit.UnitImage{
		ID:     uuid.New(),
		UnitID: unitID,
		URL:    req.URL,
	}

	result, err := h.Service.Repo.AddImage(r.Context(), image)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{
		"message": "image added successfully",
		"data":    result,
	})
}

func (h *UnitHandler) RemoveImage(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	imageID, err := uuid.Parse(r.PathValue("image_id"))
	if err != nil {
		http.Error(w, "invalid image id", http.StatusBadRequest)
		return
	}

	if claims.TenantID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
   tenantID := *claims.TenantID

	if err := h.Service.Repo.RemoveImage(r.Context(), imageID, tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}