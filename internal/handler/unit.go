package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/domain/subscription"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/domain/unit"
	"github.com/ronexlemon/bnbcore/internal/eventstream"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/upload"
	"github.com/ronexlemon/bnbcore/internal/metrics"
	"github.com/ronexlemon/bnbcore/internal/middleware"
)

type UnitHandler struct {
	Server         *http.ServeMux
	Service        *unit.UnitService
	JWTAuthManager *auth.JwtManager
	SubRepo        subscription.Repository
	Stream         *eventstream.KafkaClient
	Media         *upload.MediaService
}

func NewUnitHandler(server *http.ServeMux, service *unit.UnitService, m *auth.JwtManager,sub subscription.Repository,stream  *eventstream.KafkaClient,media *upload.MediaService,) *UnitHandler {
	h := &UnitHandler{
		Server:         server,
		Service:        service,
		JWTAuthManager: m,
		SubRepo: sub,
		Stream: stream,
		Media: media,
	}
	h.registerRoutes()
	return h
}

func (h *UnitHandler) registerRoutes() {
	api := "/api/v1"

	protected := func(hf http.HandlerFunc) http.Handler {
		return h.JWTAuthManager.Authenticate(
			middleware.RequireActiveSubscription(h.SubRepo)(metrics.MetricsMiddleware(hf)),
		)
	}
	
	h.Server.HandleFunc("GET "+api+"/units", metrics.MetricsMiddleware(h.GetAllUnits))
h.Server.Handle("GET "+api+"/host/units",
	h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.GetHostDomainDetails))))
	h.Server.HandleFunc("GET "+api+"/units/{identifier}", metrics.MetricsMiddleware(h.GetUnitByIdentifier))
	h.Server.HandleFunc("GET "+api+"/units/{id}/images", metrics.MetricsMiddleware(h.GetUnitImages))

	h.Server.Handle("POST "+api+"/units",protected(metrics.MetricsMiddleware(h.CreateUnit)))


	h.Server.Handle("PUT "+api+"/units/{id}",protected(h.UpdateUnit))

	h.Server.Handle("DELETE "+api+"/units/{id}",protected((h.DeleteUnit)))

	h.Server.Handle("POST "+api+"/units/{id}/images",protected(h.AddImage))

	h.Server.Handle("DELETE "+api+"/units/{id}/images/{image_id}",
	 h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.RemoveImage))))

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
		ShopName: result.Name,
		Location: result.Location,


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

    query := r.URL.Query()
    limit := 10
    if lStr := query.Get("limit"); lStr != "" {
        if parsedLimit, err := strconv.Atoi(lStr); err == nil && parsedLimit > 0 {
            limit = parsedLimit
        }
    }
    offset := 0
    if oStr := query.Get("offset"); oStr != "" {
        if parsedOffset, err := strconv.Atoi(oStr); err == nil && parsedOffset >= 0 {
            offset = parsedOffset
        }
    }
	units, err := h.Service.GetAllUnits(r.Context(), tenantID,limit,offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
        "data": units,
        "meta": map[string]int{
            "limit":  limit,
            "offset": offset,
        },
    })
}

func (h *UnitHandler) GetHostDomainDetails(w http.ResponseWriter, r *http.Request) {
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

    query := r.URL.Query()
    limit := 10
    if lStr := query.Get("limit"); lStr != "" {
        if parsedLimit, err := strconv.Atoi(lStr); err == nil && parsedLimit > 0 {
            limit = parsedLimit
        }
    }
    offset := 0
    if oStr := query.Get("offset"); oStr != "" {
        if parsedOffset, err := strconv.Atoi(oStr); err == nil && parsedOffset >= 0 {
            offset = parsedOffset
        }
    }
	tenantUnitsDetails, err := h.Service.GetHostUnitsDetails(r.Context(), tenantID,limit,offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
        "data": tenantUnitsDetails,
        "meta": map[string]int{
            "limit":  limit,
            "offset": offset,
        },
    })
}

func (h *UnitHandler) GetUnitImages(w http.ResponseWriter, r *http.Request) {
    unitID, err := uuid.Parse(r.PathValue("id"))
    if err != nil {
        http.Error(w, "invalid unit id", http.StatusBadRequest)
        return
    }

    t := tenant.FromContext(r.Context())
    if t == nil || t.ID == nil {
        http.Error(w, "complete workspace setup first", http.StatusPreconditionRequired)
        return
    }
    tenantID := *t.ID

    images, err := h.Service.GetUnitImages(r.Context(), unitID, tenantID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]any{
        "data": images,
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
func (h *UnitHandler) GetUnitBySlug(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		http.Error(w, "tenant not found", http.StatusBadRequest)
		return
	}

	if t.ID == nil {
		http.Error(w, "complete workspace setup first", http.StatusPreconditionRequired)
		return
	}
	tenantID := *t.ID

	slug := r.PathValue("slug")
	if slug == "" {
		http.Error(w, "invalid slug", http.StatusBadRequest)
		return
	}

	result, err := h.Service.GetBySlug(r.Context(), slug, tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]any{
		"data": result,
	})
}

func (h *UnitHandler) GetUnitByIdentifier(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil || t.ID == nil {
		http.Error(w, "tenant not found", http.StatusBadRequest)
		return
	}
	tenantID := *t.ID

	identifier := r.PathValue("identifier")
	if identifier == "" {
		http.Error(w, "missing identifier", http.StatusBadRequest)
		return
	}

	if id, err := uuid.Parse(identifier); err == nil {
		result, err := h.Service.GetUnit(r.Context(), id, tenantID)
		if err != nil {
			http.Error(w, "unit not found", http.StatusNotFound)
			return
		}

		writeJSON(w, map[string]any{"data": result})
		return
	}

	result, err := h.Service.GetBySlug(r.Context(), identifier, tenantID)
	if err != nil {
		http.Error(w, "unit not found", http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]any{"data": result})
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
	changes := map[string]any{}
	if req.Description != nil {
		changes["shop_description"] = *req.Description
	}
	if req.Name != nil {
		changes["subdomain"] = *req.Name
	}
	if req.Status != nil {
		changes["status"] = *req.Status
	}
	_ = h.Stream.Publish(r.Context(), eventstream.TopicUnitUpdated, result.ID.String(),
    eventstream.UnitUpdatedEvent{
        BaseEvent: eventstream.BaseEvent{
            TenantID:  *tenant.ID,
            UserID:    *claims.UserID,
            UserEmail: claims.Email,
            OccuredAt: time.Now(),
        },
        Title: result.Title,
		Changes: changes,
    },
)

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

	tenant:=tenant.FromContext(r.Context())

	if tenant.ID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *tenant.ID

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
	imageType := r.FormValue("image_type")
	if imageType == "" {
		http.Error(w, "missing 'image_type' field", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		http.Error(w, "file size too large or malformed form data", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing 'image' file field", http.StatusBadRequest)
		return
	}
	defer file.Close()
	cacheKey := fmt.Sprintf("unit:%s:image:%s", unitID.String(), header.Filename)
	secureURL, err := h.Media.UploadAndCacheStream(r.Context(), file, cacheKey, "hostsasa")
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to upload image stream: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Println("THE URL",secureURL)
	

	image := &unit.UnitImage{
		ID:     uuid.New(),
		UnitID: unitID,
		URL:    secureURL,
		ImageType: imageType,
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

	tenant:=tenant.FromContext(r.Context())

	if tenant.ID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *tenant.ID

	if err := h.Service.Repo.RemoveImage(r.Context(), imageID, tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}
