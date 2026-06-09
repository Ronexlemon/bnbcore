package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)


var (
	
	UserRegistrationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_registrations_total",
			Help: "Total user registration attempts.",
		},
		[]string{"status"}, // "success" | "failure"
	)

	UserLoginsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_logins_total",
			Help: "Total login attempts.",
		},
		[]string{"method", "status"}, // method: "password"|"google", status: "success"|"failure"|"unverified"
	)

	MagicLinksIssuedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "magic_links_issued_total",
			Help: "Total magic link tokens issued.",
		},
		[]string{"reason"}, // "registration" | "resend" | "login_unverified"
	)

	MagicLinksVerifiedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "magic_links_verified_total",
			Help: "Total magic link verification attempts.",
		},
		[]string{"status"}, // "success" | "invalid" | "expired"
	)

	TokenRefreshTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "token_refresh_total",
			Help: "Total token refresh attempts.",
		},
		[]string{"status"}, // "success" | "invalid" | "user_not_found"
	)

	TokenPairsIssuedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "token_pairs_issued_total",
			Help: "Total JWT access+refresh pairs successfully issued.",
		},
	)

	// Active authenticated sessions (inc on issueTokens, dec on delete/logout)
	ActiveSessionsGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_user_sessions",
			Help: "Approximate number of active user sessions (issued token pairs).",
		},
	)

	// Email sending
	VerificationEmailsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "verification_emails_total",
			Help: "Total verification emails sent.",
		},
		[]string{"status"}, // "success" | "failure"
	)
)