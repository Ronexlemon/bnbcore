package handler

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/domain/booking"
)

const (
	StatusConfirmed = "confirmed"
	StatusCanceled  = "canceled"
)



type Config struct {
	TwilioAuthToken string 
}

const maxApprovalWindow = 5 * time.Hour

type TwilioWebhookHandler struct {
	cfg     Config
	service        *booking.BookingService
	log     *slog.Logger
	Server         *http.ServeMux
}

func NewTwilioWebhookHandler(cfg Config, service  *booking.BookingService, log *slog.Logger,server   *http.ServeMux ) *TwilioWebhookHandler {
	t:= &TwilioWebhookHandler{
		cfg:     cfg,
		service: service,
		log:     log,
		Server: server,
	}
	t.registerWebhookRoutes()

	return t

}

func (h *TwilioWebhookHandler) registerWebhookRoutes() {
	api := "/api/v1"

	h.Server.HandleFunc("GET "+api+"/webhooks/twilio/whatsapp", h.Handle)
	
}



func (h *TwilioWebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.log.Error("parse form", "err", err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if !h.validateSignature(r) {
		h.log.Warn("invalid twilio signature", "remote", r.RemoteAddr)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	raw := r.FormValue("ButtonPayload")
	if raw == "" {
		raw = r.FormValue("Body")
	}

	action, bookingID,tenantID, err := parsePayload(strings.TrimSpace(strings.ToLower(raw)))
	if err != nil {
		h.log.Warn("unrecognised payload", "raw", raw, "err", err)
		h.twimlReply(w, "Unrecognised reply. Please manage bookings from the dashboard.")
		return
	}

	h.log.Info("booking action received", "action", action, "bookingID", bookingID)

	id, err := uuid.Parse(bookingID)
	if err != nil {
		h.log.Warn("invalid booking uuid", "bookingID", bookingID)
		h.twimlReply(w, "Invalid booking reference. Please use the dashboard.")
		return
	}
	tenantid, err := uuid.Parse(tenantID)
	if err != nil {
		h.log.Warn("invalid tenant uuid", "bookingID", bookingID)
		h.twimlReply(w, "Invalid booking reference. Please use the dashboard.")
		return
	}

bookResult,err	:=h.service.GetBooking(r.Context(),id,tenantid)
if err !=nil{
	h.log.Warn("invalid tenant uuid", "bookingID", bookingID)
		h.twimlReply(w, "Invalid booking reference. Please use the dashboard.")

}


if time.Since(bookResult.CreatedAt) > maxApprovalWindow {
    h.log.Warn("action rejected: link expired", 
        "bookingID", id, 
        "created_at", bookResult.CreatedAt, 
        "elapsed", time.Since(bookResult.CreatedAt),
    )
    h.twimlReply(w, "⏱️ This action has expired. The 5-hour window to manage this booking via WhatsApp has passed. Please log in to your dashboard to manage it.")
    return
}

	result, err := h.service.UpdateStatus(r.Context(), bookResult.ID,bookResult.TenantID, booking.BookingStatus(action))
	if err != nil {
		h.log.Error("update booking status", "err", err, "bookingID", id)
		h.twimlReply(w, fmt.Sprintf(
			"Could not %s booking for %s. Please try from the dashboard.",
			action, bookingID,
		))
		return
	}

	h.twimlReply(w, buildReply(action, result))
}

func buildReply(action string, b *booking.Booking) string {
	switch action {
	case StatusConfirmed:
		return fmt.Sprintf(
			"✅ Booking confirmed.\nGuest: %s\nID: %s",
			b.GuestName, b.ID,
		)
	case StatusCanceled:
		return fmt.Sprintf(
			"❌ Booking cancelled.\nGuest: %s\nID: %s",
			b.GuestName, b.ID,
		)
	default:
		return fmt.Sprintf("Booking %s updated to %s.", b.ID, action)
	}
}

func parsePayload(payload string) (action, bookingID, tenantID string, err error) {
    parts := strings.SplitN(payload, ":", 3)
    if len(parts) != 3 { 
        return "", "", "", fmt.Errorf("expected action:bookingID:tenantID, got %q", payload)
    }
    action = strings.TrimSpace(parts[0])
    bookingID = strings.TrimSpace(parts[1])
    tenantID = strings.TrimSpace(parts[2])

    if action != StatusConfirmed && action != StatusCanceled {
        return "", "", "", fmt.Errorf("unknown action %q", action)
    }
    if bookingID == "" || tenantID == "" {
        return "", "", "", fmt.Errorf("empty booking or tenant id")
    }
    return action, bookingID, tenantID, nil
}

// validateSignature checks the X-Twilio-Signature header against your auth token.
// https://www.twilio.com/docs/usage/security#validating-signatures-from-twilio
func (h *TwilioWebhookHandler) validateSignature(r *http.Request) bool {
	sig := r.Header.Get("X-Twilio-Signature")
	if sig == "" {
		return false
	}

	scheme := "https"
	if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
		scheme = "http"
	}
	fullURL := fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	keys := make([]string, 0, len(r.PostForm))
	for k := range r.PostForm {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString(fullURL)
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(url.QueryEscape(r.PostForm.Get(k)))
	}

	mac := hmac.New(sha1.New, []byte(h.cfg.TwilioAuthToken))
	mac.Write([]byte(sb.String()))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(sig))
}


func (h *TwilioWebhookHandler) twimlReply(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprintf(w,
		`<?xml version="1.0" encoding="UTF-8"?><Response><Message>%s</Message></Response>`,
		message,
	)
}