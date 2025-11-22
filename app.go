package main

import (
	"database/sql"
	"rgs/handlers"
	"rgs/middleware"
	"rgs/services"
	"rgs/sqlc"

	"github.com/go-chi/chi/v5"
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

	// Services (business logic)
	sessionsSvc := services.NewSessionsService(queries, eventBus)
	betAgg := services.NewBetAggregate(queries, walletClient, eventBus, db)
	webhookSvc := services.NewWebhookService(queries)
	outboxSvc := services.NewOutboxService(queries)

	// Handlers
	sessionsHandler := handlers.NewSessionsHandler(sessionsSvc)
	webhookHandler := handlers.NewWebhookHandler(webhookSvc)
	outboxHandler := handlers.NewOutboxHandler(outboxSvc)
	betsHandler := handlers.NewBetsHandler(betAgg)
	roundsHandler := handlers.NewRoundsHandler(queries)
	sseHandler := handlers.NewSSEHandler(eventBus)

	// Router
	r := chi.NewRouter()
	opMiddleware := middleware.NewOperatorMiddleware(queries)
	r.Use(opMiddleware.Handle)

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

	return r
}
