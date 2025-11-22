package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"rgs/middleware"
	"rgs/services"
)

type OutboxHandler struct {
	svc *services.OutboxService
}

func NewOutboxHandler(svc *services.OutboxService) *OutboxHandler {
	return &OutboxHandler{svc: svc}
}

func (h *OutboxHandler) ListOutbox(w http.ResponseWriter, r *http.Request) {
	operator, ok := middleware.OperatorFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	statusQuery := r.URL.Query().Get("status")
	var status *string
	if statusQuery != "" {
		status = &statusQuery
	}

	events, err := h.svc.ListOutbox(r.Context(), operator.ID, status)
	if err != nil {
		http.Error(w, "failed to load outbox events", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(events)
	if err != nil {
		log.Println("failed to encode outbox events")
		return
	}
}
