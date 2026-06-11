// worker/booking_notification_worker.go
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/ronexlemon/bnbcore/internal/domain/notification"
	"github.com/ronexlemon/bnbcore/internal/eventstream"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/queue"
	
)

type WhatsAppConfig struct {
	AccountSID string
	AuthToken  string
	FromNumber string
	TemplateSID  string
}

type BookingNotificationWorker struct {
	Stream   *eventstream.KafkaClient
	
	NotificationService *notification.Service
	
	Enqueuer  *queue.Enqueuer
}

func NewBookingNotificationWorker(stream *eventstream.KafkaClient,notification_service *notification.Service,enqueuer *queue.Enqueuer) *BookingNotificationWorker {
	return &BookingNotificationWorker{
		Stream:   stream,
		
		NotificationService: notification_service,
		
		Enqueuer: enqueuer,
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
			w.enqueueBookingCreated(ctx, event)

		case eventstream.TopicBookingCanceled, eventstream.TopicBookingConfirmed:
			var event eventstream.BookingStatusEvent
			if err := json.Unmarshal(value, &event); err != nil {
				log.Printf("failed to unmarshal booking status event: %v", err)
				return
			}
			w.enqueueBookingStatus(ctx, event)

		default:
			log.Printf("unhandled topic: %s", topic)
		}
	})
}

// func (w *BookingNotificationWorker) handleBookingCreated(ctx context.Context, event eventstream.BookingCreatedEvent) {


// 	notif, err := w.NotificationService.Create(ctx, notification.CreateNotificationRequest{
// 		TenantID: &event.TenantID,
// 		Type:     notification.TypeBookingCreated,
// 		Channel:  notification.ChannelWhatsApp,
// 		Title:    "New Booking Received",
// 		Message: fmt.Sprintf("Hi! New booking from %s for %s (%s - %s)",
// 			event.GuestName,
// 			event.UnitTitle,
// 			event.StartDate.Format("Jan 02"),
// 			event.EndDate.Format("Jan 02"),
// 		),
// 		Metadata: map[string]any{
// 			"booking_id":  event.BookingID,
// 			"unit_id":     event.UnitID,
// 			"guest_name":  event.GuestName,
// 			"guest_phone": event.GuestPhone,
// 			"total_price": event.TotalPrice,
// 		},
// 	})
// 	if err != nil {
// 		log.Printf("failed to save notification: %v", err)
// 	}
// 	if event.GuestPhone == "" {
// 		log.Printf("no phone number for booking %s, skipping WhatsApp notification", event.BookingID)
// 		return
// 	}
// variables := map[string]string{
//         "1": event.GuestName,
//         "2": event.UnitTitle,
//         "3": event.StartDate.Format("Jan 02, 2006"),
//         "4": event.EndDate.Format("Jan 02, 2006"),
//         "5": fmt.Sprintf("%.2f", event.TotalPrice),
//         "6": event.BookingID.String(),          
//         "7": event.TenantID.String(),    
//     }
	
// 	if err := w.TwilioClient.SendWhatsAppTemplate(ctx, event.GuestPhone, variables); err != nil {
//         log.Printf("failed to send WhatsApp for booking %s: %v", event.BookingID, err)
//         _ = w.NotificationService.MarkAsFailed(ctx, notif.ID, err.Error())
//         return
//     }
// 	_ = w.NotificationService.MarkAsSent(ctx, notif.ID)

// 	log.Printf("WhatsApp notification sent for booking %s to %s", event.BookingID, event.GuestPhone)
// }

// func (w *BookingNotificationWorker) handleBookingStatus(ctx context.Context, event eventstream.BookingStatusEvent) {
// 	if event.GuestPhone == "" {
// 		return
// 	}

// 	var message string
// 	switch event.Status {
// 	case "confirmed":
// 		message = fmt.Sprintf(
// 			"Hi %s! ✅ Your booking at *%s* has been *confirmed*.\n"+
// 				"We look forward to hosting you!",
// 			event.GuestName, event.UnitTitle,
// 		)
// 	case "canceled":
// 		message = fmt.Sprintf(
// 			"Hi %s, your booking at *%s* has been *canceled*.\n"+
// 				"Please contact us if you have any questions.",
// 			event.GuestName, event.UnitTitle,
// 		)
// 	default:
// 		return
// 	}

// 	if err := w.TwilioClient.SendWhatsApp(ctx, event.GuestPhone, message); err != nil {
// 		log.Printf("failed to send WhatsApp status update for booking %s: %v", event.BookingID, err)
// 	}
// }



//new
func (w *BookingNotificationWorker) enqueueBookingCreated(ctx context.Context, event eventstream.BookingCreatedEvent) {
	if event.GuestPhone == "" {
		log.Printf("[consumer] no phone for booking %s, skipping", event.BookingID)
		return
	}

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
		log.Printf("[consumer] failed to save notification for booking %s: %v", event.BookingID, err)
		return
	}

	payload := BookingCreatedPayload{
		NotifID:    notif.ID,
		BookingID:  event.BookingID,
		TenantID:   event.TenantID,
		GuestName:  event.GuestName,
		GuestPhone: event.GuestPhone,
		UnitTitle:  event.UnitTitle,
		StartDate:  event.StartDate,
		EndDate:    event.EndDate,
		TotalPrice: event.TotalPrice,
	}

	if _, err := queue.EnqueueTask(ctx, w.Enqueuer, WhatsAppCreatedTask, payload); err != nil {
		log.Printf("[consumer] failed to enqueue created task for booking %s: %v", event.BookingID, err)
		return
	}

	log.Printf("[consumer] job enqueued for booking %s", event.BookingID)
}

func (w *BookingNotificationWorker) enqueueBookingStatus(ctx context.Context, event eventstream.BookingStatusEvent) {
	if event.GuestPhone == "" {
		return
	}
	if event.Status != "confirmed" && event.Status != "canceled" {
		log.Printf("[consumer] ignoring unknown status %q for booking %s", event.Status, event.BookingID)
		return
	}

	payload := BookingStatusPayload{
		BookingID:  event.BookingID,
		GuestName:  event.GuestName,
		GuestPhone: event.GuestPhone,
		UnitTitle:  event.UnitTitle,
		Status:     event.Status,
	}

	if _, err := queue.EnqueueTask(ctx, w.Enqueuer, WhatsAppStatusTask, payload); err != nil {
		log.Printf("[consumer] failed to enqueue status task for booking %s: %v", event.BookingID, err)
		return
	}

	log.Printf("[consumer] status job enqueued for booking %s (%s)", event.BookingID, event.Status)
}