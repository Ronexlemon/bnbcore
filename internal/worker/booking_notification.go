// worker/booking_notification_worker.go
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"github.com/ronexlemon/bnbcore/internal/domain/notification"
	"github.com/ronexlemon/bnbcore/internal/eventstream"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/twilio"
)

type WhatsAppConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	TemplateSID  string
}

type BookingNotificationWorker struct {
	Stream   *eventstream.KafkaClient
	WhatsApp WhatsAppConfig
	NotificationService *notification.Service
	TwilioClient *twilio.Client
}

func NewBookingNotificationWorker(stream *eventstream.KafkaClient, wa WhatsAppConfig,notification_service *notification.Service,client *twilio.Client) *BookingNotificationWorker {
	return &BookingNotificationWorker{
		Stream:   stream,
		WhatsApp: wa,
		NotificationService: notification_service,
		TwilioClient: client,
	}
}

func (w *BookingNotificationWorker) Start(ctx context.Context) error {
	topics := []string{
		eventstream.TopicBookingCreated,
		eventstream.TopicBookingCanceled,
		eventstream.TopicBookingConfirmed,
	}

	return w.Stream.ConsumeEvents(ctx, topics, "booking-notification-group", func(topic string, key, value []byte) {
		switch topic {
		case eventstream.TopicBookingCreated:
			var event eventstream.BookingCreatedEvent
			if err := json.Unmarshal(value, &event); err != nil {
				log.Printf("failed to unmarshal booking created event: %v", err)
				return
			}
			fmt.Println("EVENT INCOMING", event)
			w.handleBookingCreated(ctx, event)

		case eventstream.TopicBookingCanceled, eventstream.TopicBookingConfirmed:
			var event eventstream.BookingStatusEvent
			if err := json.Unmarshal(value, &event); err != nil {
				log.Printf("failed to unmarshal booking status event: %v", err)
				return
			}
			w.handleBookingStatus(ctx, event)

		default:
			log.Printf("unhandled topic: %s", topic)
		}
	})
}

func (w *BookingNotificationWorker) handleBookingCreated(ctx context.Context, event eventstream.BookingCreatedEvent) {


	notif, err := w.NotificationService.Create(ctx, notification.CreateNotificationRequest{
		TenantID: &event.TenantID,
		Type:     notification.TypeBookingCreated,
		Channel:  notification.ChannelWhatsApp,
		Title:    "New Booking Received",
		Message: fmt.Sprintf("Hi! New booking from %s for %s (%s - %s)",
			event.GuestName,
			event.UnitTitle,
			event.StartDate.Format("Jan 02"),
			event.EndDate.Format("Jan 02"),
		),
		Metadata: map[string]any{
			"booking_id":  event.BookingID,
			"unit_id":     event.UnitID,
			"guest_name":  event.GuestName,
			"guest_phone": event.GuestPhone,
			"total_price": event.TotalPrice,
		},
	})
	if err != nil {
		log.Printf("failed to save notification: %v", err)
	}
	if event.GuestPhone == "" {
		log.Printf("no phone number for booking %s, skipping WhatsApp notification", event.BookingID)
		return
	}
variables := map[string]string{
        "1": event.GuestName,
        "2": event.UnitTitle,
        "3": event.StartDate.Format("Jan 02, 2006"),
        "4": event.EndDate.Format("Jan 02, 2006"),
        "5": fmt.Sprintf("%.2f", event.TotalPrice),
        "6": event.BookingID.String(),          
        "7": event.TenantID.String(),    
    }
	
	if err := w.TwilioClient.SendWhatsAppTemplate(ctx, event.GuestPhone, variables); err != nil {
        log.Printf("failed to send WhatsApp for booking %s: %v", event.BookingID, err)
        _ = w.NotificationService.MarkAsFailed(ctx, notif.ID, err.Error())
        return
    }
	_ = w.NotificationService.MarkAsSent(ctx, notif.ID)

	log.Printf("WhatsApp notification sent for booking %s to %s", event.BookingID, event.GuestPhone)
}

func (w *BookingNotificationWorker) handleBookingStatus(ctx context.Context, event eventstream.BookingStatusEvent) {
	if event.GuestPhone == "" {
		return
	}

	var message string
	switch event.Status {
	case "confirmed":
		message = fmt.Sprintf(
			"Hi %s! ✅ Your booking at *%s* has been *confirmed*.\n"+
				"We look forward to hosting you!",
			event.GuestName, event.UnitTitle,
		)
	case "canceled":
		message = fmt.Sprintf(
			"Hi %s, your booking at *%s* has been *canceled*.\n"+
				"Please contact us if you have any questions.",
			event.GuestName, event.UnitTitle,
		)
	default:
		return
	}

	if err := w.TwilioClient.SendWhatsApp(ctx, event.GuestPhone, message); err != nil {
		log.Printf("failed to send WhatsApp status update for booking %s: %v", event.BookingID, err)
	}
}

