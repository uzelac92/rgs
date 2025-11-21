package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"rgs/services"

	"github.com/google/uuid"
)

type SessionsReadHandler struct {
	svc *services.SessionsService
}

func NewSessionsReadHandler(svc *services.SessionsService) *SessionsReadHandler {
	return &SessionsReadHandler{svc: svc}
}

func (h *SessionsReadHandler) VerifySession(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	session, err := h.svc.VerifySession(context.Background(), id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(session)
	if err != nil {
		log.Println("Error writing session verification response")
	}
}
