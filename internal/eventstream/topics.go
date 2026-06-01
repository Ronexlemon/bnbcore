package eventstream

const (
    TopicBookingCreated  = "booking.created"
    TopicUnitCreated     = "unit.created"
    TopicBookingCanceled = "booking.canceled"
    TopicBookingConfirmed = "booking.confirmed"

	TopicUnitUpdated = "unit.updated"
	TopicUnitDeleted = "unit.deleted"

	// Tenant
	TopicTenantCreated = "tenant.created"
	TopicTenantUpdated = "tenant.updated"
	TopicTenantDeleted = "tenant.deleted"

	// Subscription
	TopicSubscriptionCreated  = "subscription.created"
	TopicSubscriptionExpiring = "subscription.expiring"
	TopicSubscriptionExpired  = "subscription.expired"
	TopicSubscriptionCanceled = "subscription.canceled"
)