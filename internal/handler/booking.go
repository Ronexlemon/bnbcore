package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/domain/booking"
	"github.com/ronexlemon/bnbcore/internal/domain/subscription"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/eventstream"
	"github.com/ronexlemon/bnbcore/internal/middleware"
)

type BookingHandler struct {
	Server         *http.ServeMux
	Service        *booking.BookingService
	JWTAuthManager *auth.JwtManager
	 Stream         *eventstream.KafkaClient
	 SubRepo        subscription.Repository
}

func NewBookingHandler(server *http.ServeMux, service *booking.BookingService, m *auth.JwtManager,stream *eventstream.KafkaClient,sub subscription.Repository ) *BookingHandler {
	h := &BookingHandler{
		Server:         server,
		Service:        service,
		JWTAuthManager: m,
		Stream: stream,
		SubRepo: sub,
	}
	h.registerRoutes()
	return h
}

func (h *BookingHandler) registerRoutes() {
	api := "/api/v1"

	 protected := func(hf http.HandlerFunc) http.Handler {
		return middleware.RequireActiveSubscription(h.SubRepo)(hf)
	}
	h.Server.HandleFunc("GET "+api+"/units/{id}/availability", h.CheckAvailability)
	h.Server.Handle("POST "+api+"/bookings",protected(h.CreateBooking))
	h.Server.HandleFunc("GET "+api+"/units/{id}/booked-dates", h.GetBookedDates)

	h.Server.Handle("GET "+api+"/bookings",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(h.GetAllBookings)))
		

	h.Server.Handle("GET "+api+"/bookings/{id}",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(h.GetBooking)))

	h.Server.Handle("GET "+api+"/units/{id}/bookings",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(h.GetBookingsByUnit)))

	h.Server.Handle("PATCH "+api+"/bookings/{id}/status",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(h.UpdateStatus)))

	h.Server.Handle("PATCH "+api+"/bookings/{id}/cancel",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(h.CancelBooking)))
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	t := tenant.FromContext(r.Context())
	if t == nil {
		http.Error(w, "tenant not found", http.StatusBadRequest)
		return
	}

	var req booking.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if t.ID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
tenantID := *t.ID
	result, err := h.Service.CreateBooking(r.Context(), tenantID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.Stream.Publish(r.Context(), eventstream.TopicBookingCreated, result.ID.String(),
        eventstream.BookingCreatedEvent{
            BookingID:  result.ID,
            TenantID:   result.TenantID,
            UnitID:     result.UnitID,
            GuestName:  result.GuestName,
            GuestEmail: result.GuestEmail,
            GuestPhone: result.GuestPhone,
            StartDate:  result.StartDate,
            EndDate:    result.EndDate,
            TotalPrice: result.TotalPrice,
            ShopName:   t.ShopDescription,
            CreatedAt:  result.CreatedAt,
        },
    )

	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]any{
		"message": "booking created successfully",
		"data":    result,
	})
}

func (h *BookingHandler) GetAllBookings(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	if claims.TenantID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
    tenantID := *claims.TenantID 

	bookings, err := h.Service.GetAllBookings(r.Context(), tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"data": bookings,
	})
}

func (h *BookingHandler) GetBookedDates(w http.ResponseWriter, r *http.Request) {
	unitID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid unit id", http.StatusBadRequest)
		return
	}

	dates, err := h.Service.GetBookedDates(r.Context(), unitID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"unit_id": unitID,
		"booked_dates": dates,
	})
}
func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid booking id", http.StatusBadRequest)
		return
	}

	if claims.TenantID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
   tenantID := *claims.TenantID

	result, err := h.Service.GetBooking(r.Context(), id, tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	writeJSON(w, map[string]any{
		"data": result,
	})
}

func (h *BookingHandler) GetBookingsByUnit(w http.ResponseWriter, r *http.Request) {
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

	if claims.TenantID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
  tenantID := *claims.TenantID

	bookings, err := h.Service.GetBookingsByUnit(r.Context(), unitID, tenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"data": bookings,
	})
}

func (h *BookingHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid booking id", http.StatusBadRequest)
		return
	}

	var req struct {
		Status booking.BookingStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Status == "" {
		http.Error(w, "status is required", http.StatusBadRequest)
		return
	}

	if claims.TenantID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
   tenantID := *claims.TenantID

	result, err := h.Service.UpdateStatus(r.Context(), id, tenantID, req.Status)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}


	if req.Status == booking.BookingStatusConfirmed || req.Status == booking.BookingStatusCanceled {
        topic := eventstream.TopicBookingConfirmed
        if req.Status == booking.BookingStatusCanceled {
            topic = eventstream.TopicBookingCanceled
        }
        _ = h.Stream.Publish(r.Context(), topic, result.ID.String(),
            eventstream.BookingStatusEvent{
                BookingID:  result.ID,
                TenantID:   result.TenantID,
                GuestName:  result.GuestName,
                GuestPhone: result.GuestPhone,
                Status:     string(req.Status),
            },
        )
    }

	writeJSON(w, map[string]any{
		"message": "booking status updated",
		"data":    result,
	})
}

func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid booking id", http.StatusBadRequest)
		return
	}

	if claims.TenantID == nil {
        http.Error(w,"complete workspace setup first" ,http.StatusPreconditionRequired)
        return
    }
   tenantID := *claims.TenantID

	if err := h.Service.CancelBooking(r.Context(), id, tenantID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]any{
		"message": "booking canceled successfully",
	})
}

func (h *BookingHandler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	unitID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid unit id", http.StatusBadRequest)
		return
	}

	startStr := r.URL.Query().Get("start_date")
	endStr := r.URL.Query().Get("end_date")
	if startStr == "" || endStr == "" {
		http.Error(w, "start_date and end_date query params are required", http.StatusBadRequest)
		return
	}

	start, err := parseDate(startStr)
	if err != nil {
		http.Error(w, "invalid start_date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	end, err := parseDate(endStr)
	if err != nil {
		http.Error(w, "invalid end_date format, use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	available, err := h.Service.Repo.CheckAvailability(r.Context(), unitID, start, end)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"unit_id":   unitID,
		"available": available,
	})
}

