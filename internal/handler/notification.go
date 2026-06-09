package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/auth"
	"github.com/ronexlemon/bnbcore/internal/domain/notification"
	"github.com/ronexlemon/bnbcore/internal/metrics"
)

type NotificationHandler struct {
	Server         *http.ServeMux
	Service        *notification.Service
	JWTAuthManager *auth.JwtManager
}

func NewNotificationHandler(server *http.ServeMux, service *notification.Service, m *auth.JwtManager) *NotificationHandler {
	h := &NotificationHandler{
		Server:         server,
		Service:        service,
		JWTAuthManager: m,
	}
	h.registerRoutes()
	return h
}

func (h *NotificationHandler) registerRoutes() {
	api := "/api/v1"

	h.Server.Handle("GET "+api+"/notifications",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.GetMyNotifications))))

	h.Server.Handle("GET "+api+"/notifications/unread",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.GetUnread))))

	h.Server.Handle("GET "+api+"/notifications/unread/count",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.UnreadCount))))

	h.Server.Handle("PATCH "+api+"/notifications/{id}/read",
		h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.MarkAsRead))))
		h.Server.Handle("POST "+api+"/notifications/read-all",
        h.JWTAuthManager.Authenticate(http.HandlerFunc(metrics.MetricsMiddleware(h.MarkAllAsRead))))

}


func (h *NotificationHandler) GetMyNotifications(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID.String())
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	limit, offset := parsePagination(r)

	notifications, err := h.Service.GetMyNotifications(r.Context(), userID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	count, _ := h.Service.UnreadCount(r.Context(), userID)

	writeJSON(w, map[string]any{
		"data":         notifications,
		"unread_count": count,
		"limit":        limit,
		"offset":       offset,
	})
}


func (h *NotificationHandler) GetUnread(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID.String())
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	notifications, err := h.Service.GetUnread(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"data":  notifications,
		"total": len(notifications),
	})
}


func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(claims.UserID.String())
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	count, err := h.Service.UnreadCount(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"unread_count": count,
	})
}


func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	claims := auth.ClaimsFromContext(r.Context())
	if claims == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "invalid notification id", http.StatusBadRequest)
		return
	}

	if err := h.Service.MarkAsRead(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]any{
		"message": "notification marked as read",
	})
}

func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
    claims := auth.ClaimsFromContext(r.Context())
    if claims == nil {
        http.Error(w, "unauthorized", http.StatusUnauthorized)
        return
    }

    userID, err := uuid.Parse(claims.UserID.String())
    if err != nil {
        http.Error(w, "invalid user id", http.StatusBadRequest)
        return
    }

    if err := h.Service.MarkAllAsRead(r.Context(), userID); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    writeJSON(w, map[string]any{
        "message": "all notifications successfully marked as read",
    })
}


func parsePagination(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0

	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return
}