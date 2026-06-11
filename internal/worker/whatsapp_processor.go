package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/ronexlemon/bnbcore/internal/domain/notification"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/queue"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/twilio"
)

type WhatsAppProcessor struct {
	TwilioClient        *twilio.Client
	NotificationService *notification.Service
}

func NewWhatsAppProcessor(
	twilioClient *twilio.Client,
	notificationService *notification.Service,
) *WhatsAppProcessor {
	return &WhatsAppProcessor{
		TwilioClient:        twilioClient,
		NotificationService: notificationService,
	}
}

// Register wires both handlers onto the asynq mux via the reusable

func (p *WhatsAppProcessor) Register(mux *asynq.ServeMux) {
	queue.RegisterHandler(mux, WhatsAppCreatedTask, p.processBookingCreated)
	queue.RegisterHandler(mux, WhatsAppStatusTask, p.processBookingStatus)
}

func (p *WhatsAppProcessor) processBookingCreated(ctx context.Context, payload BookingCreatedPayload) error {
	log.Printf("[processor] picked up created job for booking %s", payload.BookingID)

	variables := map[string]string{
		"1": payload.GuestName,
		"2": payload.UnitTitle,
		"3": payload.StartDate.Format("Jan 02, 2006"),
		"4": payload.EndDate.Format("Jan 02, 2006"),
		"5": fmt.Sprintf("%.2f", payload.TotalPrice),
		"6": payload.BookingID.String(),
		"7": payload.TenantID.String(),
	}

	log.Printf("[processor] calling Twilio for %s (booking %s)", payload.GuestPhone, payload.BookingID)

	if err := p.TwilioClient.SendWhatsAppTemplate(ctx, payload.GuestPhone, variables); err != nil {
		_ = p.NotificationService.MarkAsFailed(ctx, payload.NotifID, err.Error())
		return fmt.Errorf("twilio send failed (booking %s): %w", payload.BookingID, err)
	}

	_ = p.NotificationService.MarkAsSent(ctx, payload.NotifID)
	log.Printf("[processor] WhatsApp sent for booking %s", payload.BookingID)
	return nil
}

func (p *WhatsAppProcessor) processBookingStatus(ctx context.Context, payload BookingStatusPayload) error {
	log.Printf("[processor] picked up status job for booking %s (%s)", payload.BookingID, payload.Status)

	message, err := buildStatusMessage(payload)
	if err != nil {
		log.Printf("[processor] %v, discarding", err)
		return nil
	}

	log.Printf("[processor] calling Twilio for %s (booking %s)", payload.GuestPhone, payload.BookingID)

	if err := p.TwilioClient.SendWhatsApp(ctx, payload.GuestPhone, message); err != nil {
		return fmt.Errorf("twilio status send failed (booking %s): %w", payload.BookingID, err)
	}

	log.Printf("[processor] WhatsApp status sent for booking %s (%s)", payload.BookingID, payload.Status)
	return nil
}

func buildStatusMessage(payload BookingStatusPayload) (string, error) {
	switch payload.Status {
	case "confirmed":
		return fmt.Sprintf(
			"Hi %s! ✅ Your booking at *%s* has been *confirmed*.\nWe look forward to hosting you!",
			payload.GuestName, payload.UnitTitle,
		), nil
	case "canceled":
		return fmt.Sprintf(
			"Hi %s, your booking at *%s* has been *canceled*.\nPlease contact us if you have any questions.",
			payload.GuestName, payload.UnitTitle,
		), nil
	default:
		return "", fmt.Errorf("unknown status %q for booking %s", payload.Status, payload.BookingID)
	}
}