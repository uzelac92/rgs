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

type SessionsHandler struct {
	svc *services.SessionsService
}

func NewSessionsHandler(svc *services.SessionsService) *SessionsHandler {
	return &SessionsHandler{svc: svc}
}

type launchRequest struct {
	ExternalPlayerID string `json:"external_player_id"`
	Jurisdiction     string `json:"jurisdiction"`
	TTL              int32  `json:"ttl_seconds"`
}

type launchResponse struct {
	SessionToken string    `json:"session_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func (h *SessionsHandler) VerifySession(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("sess_token")
	if token == "" {
		http.Error(w, "missing sess_token", http.StatusBadRequest)
		return
	}

	operator, ok := middleware.OperatorFromContext(r.Context())
	if !ok {
		http.Error(w, "missing operator context", http.StatusUnauthorized)
		return
	}

	session, err := h.svc.VerifySession(r.Context(), token, operator.ID)
	if err != nil {
		http.Error(w, "invalid or expired session", http.StatusUnauthorized)
		return
	}

	if errEncode := json.NewEncoder(w).Encode(session); errEncode != nil {
		log.Println("failed to write session response:", errEncode)
	}
}

func (h *SessionsHandler) LaunchSession(w http.ResponseWriter, r *http.Request) {
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

	res := launchResponse{
		SessionToken: session.LaunchToken,
		ExpiresAt:    session.ExpiresAt,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		log.Println("error encoding session response:", err)
	}
}

func (h *SessionsHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	operator, ok := middleware.OperatorFromContext(r.Context())
	if !ok {
		http.Error(w, "operator missing from context", http.StatusUnauthorized)
		return
	}

	if err := h.svc.RevokeSession(context.Background(), id, operator.ID); err != nil {
		http.Error(w, "failed to revoke", http.StatusInternalServerError)
		return
	}

	_, err = w.Write([]byte("revoked"))
	if err != nil {
		log.Println("Error writing revocation response")
		return
	}
}
