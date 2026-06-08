package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/domain/booking"
	"github.com/ronexlemon/bnbcore/internal/domain/services"
	"github.com/ronexlemon/bnbcore/internal/domain/tenant"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/twilio"
)



type CheckoutNotificationWorker struct {
	bookingService            *booking.BookingService
	twilioClient    *twilio.Client
	cleanerService  *services.Service
	tenantService   *tenant.Service
	batchSize       int
	interBatchDelay time.Duration
}

func NewCheckoutNotificationWorker(
	bs *booking.BookingService,
	tc *twilio.Client,
	cs  *services.Service,
	ts *tenant.Service,
	batchSize int,
	delay time.Duration,
) *CheckoutNotificationWorker {
	return &CheckoutNotificationWorker{
		bookingService:            bs,
		twilioClient:    tc,
		cleanerService:  cs,
		tenantService:   ts,
		batchSize:       batchSize,
		interBatchDelay: delay,
	}
}

func (w *CheckoutNotificationWorker) Start(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return

        case <-ticker.C:
              w.ProcessTodayCheckouts(ctx)
                log.Printf("worker error: %v")
            
        }
    }
}

func (w *CheckoutNotificationWorker) ProcessTodayCheckouts(ctx context.Context) {
	today := time.Now().UTC()
	log.Printf("[Checkout Worker] Executing notifications loop for date: %s", today.Format("2006-01-02"))

	var lastID uuid.UUID = uuid.Nil
	totalProcessed := 0

	for {
		bookings, err := w.bookingService.FindConfirmedBookingsEndingOnDate(ctx, today, lastID, w.batchSize)
		log.Printf("[Checkout Worker] Checking for date: %s", today.Format("2006-01-02"))
		if err != nil {
			log.Printf("[Checkout Worker] Database recovery error: %v", err)
			return
		}

		if len(bookings) == 0 {
			break 
		}

		for _, b := range bookings {
			
			bookingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			w.notifyAllParties(bookingCtx, b)
			cancel()

			lastID = b.ID
			totalProcessed++
		}

		if w.interBatchDelay > 0 {
			time.Sleep(w.interBatchDelay)
		}
	}

	log.Printf("[Checkout Worker] Completed job run. Total bookings dispatched: %d", totalProcessed)
}


func (w *CheckoutNotificationWorker) notifyAllParties(ctx context.Context, b *booking.Booking) {
	log.Printf("[Checkout Worker] Processing notifications for Booking ID: %s", b.ID)


	if b.GuestPhone != "" {
		guestMsg := fmt.Sprintf(
			"Hi %s! 👋 Just a reminder that your checkout is today. We hope you enjoyed your stay! Please let us know once you have locked up.",
			b.GuestName,
		)
		if err := w.twilioClient.SendWhatsApp(ctx, b.GuestPhone, guestMsg); err != nil {
			log.Printf("Err: failed to notify Guest (%s) via WhatsApp: %v", b.ID, err)
		}
	}

	
	 tenant, err := w.tenantService.GetTenantByID(ctx, b.TenantID)
	if err != nil {
		log.Printf("Err: failed to resolve tenant phone for tenant %s: %v", b.TenantID, err)
	} else if tenant.PhoneNumber != "" { 
		tenantMsg := fmt.Sprintf(
			"🔔 Checkout Alert: Guest %s is scheduled to check out of your unit today (Booking ID: %s).",
			b.GuestName, b.ID,
		)
		if err := w.twilioClient.SendWhatsApp(ctx, "tenant.phoneNumber", tenantMsg); err != nil {
			log.Printf("Err: failed to notify Tenant (%s) via WhatsApp: %v", b.TenantID, err)
		}
	}

	cleanerService, err := w.cleanerService.GetByID(ctx, b.UnitID,b.TenantID)
	if err != nil {
		log.Printf("Err: failed to resolve cleaner phone for unit %s: %v", b.UnitID, err)
	} else if cleanerService.Mobile != "" {
		cleanerMsg := fmt.Sprintf(
			"🧹 Cleaning Task: A checkout is occurring today for Unit reference: %s. The property will be ready for turnover cleanup shortly.",
			b.UnitID,
		)
		if err := w.twilioClient.SendWhatsApp(ctx, cleanerService.Mobile, cleanerMsg); err != nil {
			log.Printf("Err: failed to notify Cleaner via WhatsApp: %v", err)
		}
	}
}