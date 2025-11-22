package main

import (
	"database/sql"
	"rgs/handlers"
	"rgs/middleware"
	"rgs/observability"
	"rgs/services"
	"rgs/sqlc"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func BuildApp(db *sql.DB, cfg Config) *chi.Mux {
	queries := sqlc.New(db)

	walletClient := services.NewWalletClient(
		cfg.WalletUrl,
		cfg.WalletSecret,
	)

	eventBus := services.NewEventBus(100)

	outboxWorker := services.NewOutboxWorker(queries, walletClient, eventBus)
	outboxWorker.Start()

	webhookWorker := services.NewWebhookWorker(queries, eventBus)
	webhookWorker.Start()

	complianceSvc := services.NewComplianceService(queries)

	// Services (business logic)
	sessionsSvc := services.NewSessionsService(queries, eventBus, complianceSvc)
	betAgg := services.NewBetAggregate(queries, walletClient, eventBus, db, complianceSvc)
	webhookSvc := services.NewWebhookService(queries)
	outboxSvc := services.NewOutboxService(queries)

	// Handlers
	sessionsHandler := handlers.NewSessionsHandler(sessionsSvc)
	webhookHandler := handlers.NewWebhookHandler(webhookSvc)
	outboxHandler := handlers.NewOutboxHandler(outboxSvc)
	betsHandler := handlers.NewBetsHandler(betAgg)
	roundsHandler := handlers.NewRoundsHandler(queries)
	sseHandler := handlers.NewSSEHandler(eventBus)
	auditHandler := handlers.NewAuditHandler(complianceSvc)

	// Router
	r := chi.NewRouter()
	opMiddleware := middleware.NewOperatorMiddleware(queries)
	rateLimiter := middleware.NewRateLimiter(20, 10)
	r.Use(opMiddleware.Handle)
	r.Use(rateLimiter.Limit)
	r.Use(observability.MetricsMiddleware)
	r.Use(middleware.OTelMiddleware)

	// Sessions
	r.Post("/sessions/launch", sessionsHandler.LaunchSession)
	r.Post("/sessions/revoke", sessionsHandler.RevokeSession)
	r.Get("/sessions/verify", sessionsHandler.VerifySession)

	// Bets
	r.Post("/bets", betsHandler.PlaceBet)

	// Rounds
	r.Get("/rounds/{id}", roundsHandler.GetRound)

	// Webhooks
	r.Get("/webhooks", webhookHandler.ListWebhooks)
	r.Post("/webhooks/retry/{id}", webhookHandler.RetryWebhook)

	// Outbox
	r.Get("/outbox", outboxHandler.ListOutbox)

	// Stream
	r.Get("/stream", sseHandler.Stream)

	// Audit
	r.Get("/audit", auditHandler.List)

	// Metrics
	r.Handle("/metrics", promhttp.Handler())

	return r
}
