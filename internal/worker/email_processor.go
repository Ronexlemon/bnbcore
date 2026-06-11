package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/ronexlemon/bnbcore/internal/domain/notification"
	"github.com/ronexlemon/bnbcore/internal/infrastructure/queue"
	"github.com/ronexlemon/bnbcore/internal/metrics"
	"github.com/ronexlemon/bnbcore/internal/senders"
)

type EmailProcessor struct {
	EmailSender         *senders.Sender
	NotificationService *notification.Service
}

func NewEmailProcessor(
	emailSender *senders.Sender,
	notificationService *notification.Service,
) *EmailProcessor {
	return &EmailProcessor{
		EmailSender:         emailSender,
		NotificationService: notificationService,
	}
}

// Register wires email handlers onto the asynq mux.
func (p *EmailProcessor) Register(mux *asynq.ServeMux) {
	queue.RegisterHandler(mux, UserSignUpTask, p.processSignUpRegistration)
}

func (p *EmailProcessor) processSignUpRegistration(ctx context.Context, payload UserSignupPayload) error {
	log.Printf("[processor] sending signup email to %s (user %s)", payload.Email, payload.UserID)

	emailPayload := senders.EmailPayload{
		To:      payload.Email,
		Subject: "Welcome to Hostsasa — confirm your signup",
		Title:   "Welcome to Hostsasa",
		Greeting: "Hi,",
		Body:    fmt.Sprintf("Click the link below to confirm your account and get started:\n\n\n%s", payload.Link),
		Footer:  "If you didn't request this, you can safely ignore this email.\n\n— The Hostsasa Team",
	}

	if err := p.EmailSender.Send(emailPayload); err != nil {
		metrics.VerificationEmailsTotal.WithLabelValues("failure").Inc()
		return fmt.Errorf("signup email send failed (user %s): %w", payload.UserID, err)
	}
     metrics.MagicLinksIssuedTotal.WithLabelValues("registration").Inc()
	metrics.VerificationEmailsTotal.WithLabelValues("success").Inc()
	log.Printf("[processor] signup email sent (user %s)", payload.UserID)
	return nil
}