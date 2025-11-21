package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"rgs/middleware"
	"time"

	"rgs/services"

	"github.com/google/uuid"
)

type SessionsWriteHandler struct {
	svc *services.SessionsService
}

func NewSessionsWriteHandler(svc *services.SessionsService) *SessionsWriteHandler {
	return &SessionsWriteHandler{svc: svc}
}

type launchRequest struct {
	ExternalPlayerID string `json:"external_player_id"`
	Jurisdiction     string `json:"jurisdiction"`
	TTL              int32  `json:"ttl_seconds"`
}

func (h *SessionsWriteHandler) LaunchSession(w http.ResponseWriter, r *http.Request) {
	var req launchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	operator, ok := middleware.OperatorFromContext(r.Context())
	if !ok {
		http.Error(w, "operator missing from context", http.StatusUnauthorized)
		return
	}

	session, err := h.svc.LaunchSession(r.Context(), services.LaunchSessionParams{
		OperatorID:       operator.ID,
		ExternalPlayerID: req.ExternalPlayerID,
		Jurisdiction:     req.Jurisdiction,
		TTL:              time.Duration(req.TTL) * time.Second,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(session)
	if err != nil {
		log.Println("error encoding session launch:", err)
		return
	}
}

func (h *SessionsWriteHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	if err := h.svc.RevokeSession(context.Background(), id); err != nil {
		http.Error(w, "failed to revoke", http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte("revoked"))
	if err != nil {
		log.Println("Error writing revocation response")
		return
	}
}
