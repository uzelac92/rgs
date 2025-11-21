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

	sessionsSvc := services.NewSessionsService(q)

	sessionsWrite := handlers.NewSessionsWriteHandler(sessionsSvc)
	sessionsRead := handlers.NewSessionsReadHandler(sessionsSvc)

	r.Post("/sessions/launch", sessionsWrite.LaunchSession)
	r.Post("/sessions/revoke", sessionsWrite.RevokeSession)

	r.Get("/sessions/verify", sessionsRead.VerifySession)

	return r
}
