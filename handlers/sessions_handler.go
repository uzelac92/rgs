package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"rgs/middleware"
	"rgs/observability"
	"time"

	"rgs/services"

	"github.com/google/uuid"
	"go.uber.org/zap"
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
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	operator, ok := middleware.OperatorFromContext(r.Context())
	if !ok {
		http.Error(w, "missing operator context", http.StatusUnauthorized)
		return
	}

	observability.Logger.Info("verify session",
		zap.Int32("operator_id", operator.ID),
	)
	session, err := h.svc.VerifySession(r.Context(), token, operator.ID)
	if err != nil {
		http.Error(w, "invalid or expired session", http.StatusUnauthorized)
		observability.Logger.Error("invalid or expired session", zap.Error(err))
		return
	}

	if errEncode := json.NewEncoder(w).Encode(session); errEncode != nil {
		observability.Logger.Error("failed to write session response", zap.Error(err))
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

	observability.Logger.Info("launching session",
		zap.Int32("operator_id", operator.ID),
		zap.Any("external_player_id", req.ExternalPlayerID),
		zap.Any("jurisdiction", req.Jurisdiction),
		zap.Int32("ttl_seconds", req.TTL),
	)
	session, err := h.svc.LaunchSession(r.Context(), services.LaunchSessionParams{
		OperatorID:       operator.ID,
		ExternalPlayerID: req.ExternalPlayerID,
		Jurisdiction:     req.Jurisdiction,
		TTL:              time.Duration(req.TTL) * time.Second,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		observability.Logger.Error("session launch failed", zap.Error(err))
		return
	}

	res := launchResponse{
		SessionToken: session.LaunchToken,
		ExpiresAt:    session.ExpiresAt,
	}

	if err := json.NewEncoder(w).Encode(res); err != nil {
		observability.Logger.Error("error encoding session response", zap.Error(err))
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

	observability.Logger.Info("revoking session",
		zap.Int32("operator_id", operator.ID),
	)
	if err := h.svc.RevokeSession(context.Background(), id, operator.ID); err != nil {
		http.Error(w, "failed to revoke", http.StatusInternalServerError)
		observability.Logger.Error("session revoking failed", zap.Error(err))
		return
	}

	_, err = w.Write([]byte("revoked"))
	if err != nil {
		observability.Logger.Error("error writing revocation response", zap.Error(err))
		return
	}
}
