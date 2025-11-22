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

	// Start async outbox worker
	outboxWorker := services.NewOutboxWorker(queries, walletClient)
	outboxWorker.Start()

	// Start async webhook worker
	webhookWorker := services.NewWebhookWorker(queries)
	webhookWorker.Start()

	// Services (business logic)
	sessionsSvc := services.NewSessionsService(queries)
	betAgg := services.NewBetAggregate(queries, walletClient)
	webhookSvc := services.NewWebhookService(queries)
	outboxSvc := services.NewOutboxService(queries)

	// Handlers
	sessionsWrite := handlers.NewSessionsWriteHandler(sessionsSvc)
	sessionsRead := handlers.NewSessionsReadHandler(sessionsSvc)
	webhookHandler := handlers.NewWebhookHandler(webhookSvc)
	outboxHandler := handlers.NewOutboxHandler(outboxSvc)

	betsWrite := handlers.NewBetsWriteHandler(betAgg)
	roundsRead := handlers.NewRoundsReadHandler(queries)

	// Router
	r := chi.NewRouter()
	opMiddleware := middleware.NewOperatorMiddleware(queries)
	r.Use(opMiddleware.Handle)

	// Sessions
	r.Post("/sessions/launch", sessionsWrite.LaunchSession)
	r.Post("/sessions/revoke", sessionsWrite.RevokeSession)
	r.Get("/sessions/verify", sessionsRead.VerifySession)

	// Bets
	r.Post("/bets", betsWrite.PlaceBet)

	// Rounds
	r.Get("/rounds/{id}", roundsRead.GetRound)

	// Webhooks
	r.Get("/webhooks", webhookHandler.ListWebhooks)
	r.Post("/webhooks/retry/{id}", webhookHandler.RetryWebhook)

	// Outbox
	r.Get("/outbox", outboxHandler.ListOutbox)

	return r
}
