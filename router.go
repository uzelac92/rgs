package main

import (
	"rgs/handlers"
	"rgs/middleware"
	"rgs/services"
	"rgs/sqlc"

	"github.com/go-chi/chi/v5"
)

func SetupRouter(q *sqlc.Queries) *chi.Mux {
	r := chi.NewRouter()

	opMiddleware := middleware.NewOperatorMiddleware(q)
	r.Use(opMiddleware.Handle)

	// Sessions routing
	sessionsSvc := services.NewSessionsService(q)

	sessionsWrite := handlers.NewSessionsWriteHandler(sessionsSvc)
	sessionsRead := handlers.NewSessionsReadHandler(sessionsSvc)

	r.Post("/sessions/launch", sessionsWrite.LaunchSession)
	r.Post("/sessions/revoke", sessionsWrite.RevokeSession)

	r.Get("/sessions/verify", sessionsRead.VerifySession)

	// Bets routing
	betAgg := services.NewBetAggregate(q)
	betsWrite := handlers.NewBetsWriteHandler(betAgg)

	r.Post("/bets", betsWrite.PlaceBet)

	// Rounds routing
	roundsRead := handlers.NewRoundsReadHandler(q)
	r.Get("/rounds/{id}", roundsRead.GetRound)

	return r
}
