
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/ronexlemon/bnbcore/internal/domain/notification"
	"github.com/ronexlemon/bnbcore/internal/senders"

	"github.com/ronexlemon/bnbcore/internal/eventstream"
)

type NotificationWorker struct {
	Stream       *eventstream.KafkaClient
	NotifService *notification.Service
	EmailSender  *senders.Sender
}

func NewNotificationWorker(
	stream *eventstream.KafkaClient,
	notifService *notification.Service,
	emailSender *senders.Sender,
) *NotificationWorker {
	return &NotificationWorker{
		Stream:       stream,
		NotifService: notifService,
		EmailSender:  emailSender,
	}
}

func (w *NotificationWorker) Start(ctx context.Context) error {
	topics := []string{
		eventstream.TopicTenantCreated,
		eventstream.TopicTenantUpdated,
		eventstream.TopicTenantDeleted,
		eventstream.TopicUnitCreated,
		eventstream.TopicUnitUpdated,
		eventstream.TopicUnitDeleted,
		eventstream.TopicBookingCreated,
		eventstream.TopicBookingCanceled,
		eventstream.TopicBookingConfirmed,
		eventstream.TopicSubscriptionCreated,
		eventstream.TopicSubscriptionExpiring,
		eventstream.TopicSubscriptionExpired,
		eventstream.TopicSubscriptionCanceled,
	}
	fmt.Println("Starting on general")
	//ConsumeEvents(ctx context.Context, topics []string, groupID string, handler func(topic string, key []byte, value []byte))

	return w.Stream.ConsumeEvents(ctx, topics, "general-notification-group", func(topic string, key, value []byte) {
    log.Printf("received event on topic: %s", topic)
	


    switch topic {
    case eventstream.TopicTenantCreated:
        var e eventstream.TenantCreatedEvent
        if err := json.Unmarshal(value, &e); err != nil {
            log.Printf("failed to unmarshal tenant created: %v", err)
            return
        }
        w.onTenantCreated(ctx, e)

    case eventstream.TopicTenantUpdated:
        var e eventstream.TenantUpdatedEvent
        if err := json.Unmarshal(value, &e); err != nil {
            log.Printf("failed to unmarshal tenant updated: %v", err)
            return
        }
        w.onTenantUpdated(ctx, e)

    case eventstream.TopicTenantDeleted:
        var e eventstream.TenantDeletedEvent
        if err := json.Unmarshal(value, &e); err != nil {
            log.Printf("failed to unmarshal tenant deleted: %v", err)
            return
        }
        w.onTenantDeleted(ctx, e)

    // ── Unit ──────────────────────────────────────────────────
    case eventstream.TopicUnitCreated:
        var e eventstream.UnitCreatedEvent
        if err := json.Unmarshal(value, &e); err != nil {
            log.Printf("failed to unmarshal unit created: %v", err)
            return
        }
        w.onUnitCreated(ctx, e)

    case eventstream.TopicUnitUpdated:
        var e eventstream.UnitUpdatedEvent
        if err := json.Unmarshal(value, &e); err != nil {
            log.Printf("failed to unmarshal unit updated: %v", err)
            return
        }
        w.onUnitUpdated(ctx, e)

    case eventstream.TopicUnitDeleted:
        var e eventstream.UnitDeletedEvent
        if err := json.Unmarshal(value, &e); err != nil {
            log.Printf("failed to unmarshal unit deleted: %v", err)
            return
        }
        w.onUnitDeleted(ctx, e)

    
    // ── Subscription ──────────────────────────────────────────
    case eventstream.TopicSubscriptionCreated:
        var e eventstream.SubscriptionEvent
        if err := json.Unmarshal(value, &e); err != nil {
            log.Printf("failed to unmarshal subscription created: %v", err)
            return
        }
        w.onSubscriptionCreated(ctx, e)

    case eventstream.TopicSubscriptionExpiring:
        var e eventstream.SubscriptionEvent
        if err := json.Unmarshal(value, &e); err != nil {
            log.Printf("failed to unmarshal subscription expiring: %v", err)
            return
        }
        w.onSubscriptionExpiring(ctx, e)

    case eventstream.TopicSubscriptionExpired:
        var e eventstream.SubscriptionEvent
        if err := json.Unmarshal(value, &e); err != nil {
            log.Printf("failed to unmarshal subscription expired: %v", err)
            return
        }
        w.onSubscriptionExpired(ctx, e)

    case eventstream.TopicSubscriptionCanceled:
        var e eventstream.SubscriptionEvent
        if err := json.Unmarshal(value, &e); err != nil {
            log.Printf("failed to unmarshal subscription canceled: %v", err)
            return
        }
        w.onSubscriptionCanceled(ctx, e)
    }
})
}

// ── Tenant handlers ───────────────────────────────────────────────────────────

func (w *NotificationWorker) onTenantCreated(ctx context.Context, e eventstream.TenantCreatedEvent) {
	w.saveAndEmail(ctx, saveAndEmailParams{
		TenantID: &e.TenantID,
		UserID:   &e.UserID,
		Type:     notification.TypeUnitCreated,
		Title:    "Welcome to BnBCore! 🎉",
		Message:  fmt.Sprintf("Your shop *%s* has been created. Subdomain: %s", e.ShopName, e.Subdomain),
		Metadata: map[string]any{"subdomain": e.Subdomain},
		Email:    e.UserEmail,
		Subject:  "Welcome to BnBCore — Your shop is ready!",
		Body: senders.BuildHTML(
			"Welcome to BnBCore!",
			fmt.Sprintf("Hi there! Your shop <b>%s</b> has been created successfully.", e.ShopName),
			fmt.Sprintf("Your subdomain: <b>%s.bnbcore.com</b><br>You can now start adding your units.", e.Subdomain),
			"Thank you for choosing BnBCore.",
		),
	})
}

func (w *NotificationWorker) onTenantUpdated(ctx context.Context, e eventstream.TenantUpdatedEvent) {
	w.saveAndEmail(ctx, saveAndEmailParams{
		TenantID: &e.TenantID,
		UserID:   &e.UserID,
		Type:     notification.TypeUnitCreated,
		Title:    "Shop Updated",
		Message:  fmt.Sprintf("Your shop *%s* details have been updated.", e.ShopName),
		Metadata: e.Changes,
		Email:    e.UserEmail,
		Subject:  "Your shop has been updated",
		Body: senders.BuildHTML(
			"Shop Updated",
			fmt.Sprintf("Hi! Your shop <b>%s</b> has been updated.", e.ShopName),
			"Your shop details have been successfully updated.",
			"If you did not make this change, please contact support.",
		),
	})
}

func (w *NotificationWorker) onTenantDeleted(ctx context.Context, e eventstream.TenantDeletedEvent) {
	w.saveAndEmail(ctx, saveAndEmailParams{
		TenantID: &e.TenantID,
		UserID:   &e.UserID,
		Type:     notification.TypeUnitCreated,
		Title:    "Shop Deleted",
		Message:  fmt.Sprintf("Your shop *%s* has been deleted.", e.ShopName),
		Metadata: nil,
		Email:    e.UserEmail,
		Subject:  "Your shop has been deleted",
		Body: senders.BuildHTML(
			"Shop Deleted",
			fmt.Sprintf("Hi! Your shop <b>%s</b> has been deleted.", e.ShopName),
			"All your data has been removed. We're sorry to see you go.",
			"If this was a mistake, please contact support immediately.",
		),
	})
}

// ── Unit handlers ─────────────────────────────────────────────────────────────

func (w *NotificationWorker) onUnitCreated(ctx context.Context, e eventstream.UnitCreatedEvent) {
	w.saveAndEmail(ctx, saveAndEmailParams{
		TenantID: &e.TenantID,
		UserID:   &e.UserID,
		Type:     notification.TypeUnitCreated,
		Title:    "Unit Added ✅",
		Message:  fmt.Sprintf("Unit *%s* in %s has been added to your shop.", e.Title, e.Location),
		Metadata: map[string]any{"unit_id": e.UnitID, "title": e.Title},
		Email:    e.UserEmail,
		Subject:  "New unit added to your shop",
		Body: senders.BuildHTML(
			"Unit Added",
			fmt.Sprintf("Hi! A new unit has been added to <b>%s</b>.", e.ShopName),
			fmt.Sprintf("<b>%s</b><br>Location: %s", e.Title, e.Location),
			"Log in to manage your units.",
		),
	})
}

func (w *NotificationWorker) onUnitUpdated(ctx context.Context, e eventstream.UnitUpdatedEvent) {
	w.saveAndEmail(ctx, saveAndEmailParams{
		TenantID: &e.TenantID,
		UserID:   &e.UserID,
		Type:     notification.TypeUnitCreated,
		Title:    "Unit Updated",
		Message:  fmt.Sprintf("Unit *%s* has been updated.", e.Title),
		Metadata: e.Changes,
		Email:    e.UserEmail,
		Subject:  "Your unit has been updated",
		Body: senders.BuildHTML(
			"Unit Updated",
			fmt.Sprintf("Hi! Your unit <b>%s</b> has been updated.", e.Title),
			"The changes have been saved successfully.",
			"Log in to view the changes.",
		),
	})
}

func (w *NotificationWorker) onUnitDeleted(ctx context.Context, e eventstream.UnitDeletedEvent) {
	w.saveAndEmail(ctx, saveAndEmailParams{
		TenantID: &e.TenantID,
		UserID:   &e.UserID,
		Type:     notification.TypeUnitCreated,
		Title:    "Unit Deleted",
		Message:  fmt.Sprintf("Unit *%s* has been deleted from your shop.", e.Title),
		Metadata: map[string]any{"unit_id": e.UnitID},
		Email:    e.UserEmail,
		Subject:  "A unit has been deleted",
		Body: senders.BuildHTML(
			"Unit Deleted",
			fmt.Sprintf("Hi! Your unit <b>%s</b> has been deleted.", e.Title),
			"This unit is no longer available for booking.",
			"If this was a mistake, please contact support.",
		),
	})
}




// ── Subscription handlers ─────────────────────────────────────────────────────

func (w *NotificationWorker) onSubscriptionCreated(ctx context.Context, e eventstream.SubscriptionEvent) {
	w.saveAndEmail(ctx, saveAndEmailParams{
		TenantID: &e.TenantID,
		UserID:   &e.UserID,
		Type:     notification.TypePaymentReceived,
		Title:    "Subscription Active 🎉",
		Message:  fmt.Sprintf("You are now on the *%s* plan (%s). Expires: %s", e.Plan, e.BillingCycle, e.ExpiresAt.Format("Jan 02, 2006")),
		Metadata: map[string]any{"plan": e.Plan, "billing_cycle": e.BillingCycle, "expires_at": e.ExpiresAt},
		Email:    e.UserEmail,
		Subject:  "Subscription activated",
		Body: senders.BuildHTML(
			"Subscription Activated",
			fmt.Sprintf("Hi! Your <b>%s</b> plan is now active.", e.Plan),
			fmt.Sprintf("Billing: <b>%s</b><br>Amount: <b>%s %.2f</b><br>Expires: <b>%s</b>",
				e.BillingCycle, e.Currency, e.Amount, e.ExpiresAt.Format("Jan 02, 2006")),
			"Thank you for subscribing to BnBCore.",
		),
	})
}

func (w *NotificationWorker) onSubscriptionExpiring(ctx context.Context, e eventstream.SubscriptionEvent) {
	w.saveAndEmail(ctx, saveAndEmailParams{
		TenantID: &e.TenantID,
		UserID:   &e.UserID,
		Type:     notification.TypeSubscriptionExpiring,
		Title:    "Subscription Expiring Soon ⚠️",
		Message:  fmt.Sprintf("Your *%s* plan expires on %s. Renew to keep access.", e.Plan, e.ExpiresAt.Format("Jan 02, 2006")),
		Metadata: map[string]any{"plan": e.Plan, "expires_at": e.ExpiresAt},
		Email:    e.UserEmail,
		Subject:  "Your subscription is expiring soon",
		Body: senders.BuildHTML(
			"Subscription Expiring Soon",
			"Hi! Your subscription is about to expire.",
			fmt.Sprintf("Plan: <b>%s</b><br>Expires: <b>%s</b><br><br>Renew now to avoid losing access.", e.Plan, e.ExpiresAt.Format("Jan 02, 2006")),
			"Log in to renew your subscription.",
		),
	})
}

func (w *NotificationWorker) onSubscriptionExpired(ctx context.Context, e eventstream.SubscriptionEvent) {
	w.saveAndEmail(ctx, saveAndEmailParams{
		TenantID: &e.TenantID,
		UserID:   &e.UserID,
		Type:     notification.TypeSubscriptionExpired,
		Title:    "Subscription Expired ❌",
		Message:  fmt.Sprintf("Your *%s* plan has expired. Renew to restore access.", e.Plan),
		Metadata: map[string]any{"plan": e.Plan},
		Email:    e.UserEmail,
		Subject:  "Your subscription has expired",
		Body: senders.BuildHTML(
			"Subscription Expired",
			"Hi! Your subscription has expired.",
			fmt.Sprintf("Your <b>%s</b> plan has expired and your account has been restricted.<br><br>Please renew to restore full access.", e.Plan),
			"Log in to renew your subscription.",
		),
	})
}

func (w *NotificationWorker) onSubscriptionCanceled(ctx context.Context, e eventstream.SubscriptionEvent) {
	w.saveAndEmail(ctx, saveAndEmailParams{
		TenantID: &e.TenantID,
		UserID:   &e.UserID,
		Type:     notification.TypeSubscriptionExpired,
		Title:    "Subscription Canceled",
		Message:  fmt.Sprintf("Your *%s* plan has been canceled.", e.Plan),
		Metadata: map[string]any{"plan": e.Plan},
		Email:    e.UserEmail,
		Subject:  "Your subscription has been canceled",
		Body: senders.BuildHTML(
			"Subscription Canceled",
			"Hi! Your subscription has been canceled.",
			fmt.Sprintf("Your <b>%s</b> plan has been canceled. You will retain access until <b>%s</b>.", e.Plan, e.ExpiresAt.Format("Jan 02, 2006")),
			"We're sorry to see you go. Contact support if this was a mistake.",
		),
	})
}

// ── core helper ───────────────────────────────────────────────────────────────

type saveAndEmailParams struct {
	TenantID *uuid.UUID
	UserID   *uuid.UUID
	Type     notification.Type
	Title    string
	Message  string
	Metadata map[string]any
	Email    string
	Subject  string
	Body     string
}

func (w *NotificationWorker) saveAndEmail(ctx context.Context, p saveAndEmailParams) {
	notif, err := w.NotifService.Create(ctx, notification.CreateNotificationRequest{
		TenantID: p.TenantID,
		UserID:   p.UserID,
		Type:     p.Type,
		Channel:  notification.ChannelInApp,
		Title:    p.Title,
		Message:  p.Message,
		Metadata: p.Metadata,
	})
	if err != nil {
		log.Printf("failed to save in-app notification: %v", err)
	}

	if p.Email == "" {
		log.Printf("no email address, skipping email notification")
		return
	}

	if err := w.EmailSender.Send(senders.EmailPayload{
		To:      p.Email,
		Subject: p.Subject,
		Body:    p.Body,
	}); err != nil {
		log.Printf("failed to send email to %s: %v", p.Email, err)
		if notif != nil {
			_ = w.NotifService.MarkAsFailed(ctx, notif.ID, err.Error())
		}
		return
	}

	if notif != nil {
		_ = w.NotifService.MarkAsSent(ctx, notif.ID)
	}
	log.Printf("notification sent to %s — %s", p.Email, p.Title)
}