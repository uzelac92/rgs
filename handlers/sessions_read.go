package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"rgs/middleware"

	"rgs/services"
)

type SessionsReadHandler struct {
	svc *services.SessionsService
}

func NewSessionsReadHandler(svc *services.SessionsService) *SessionsReadHandler {
	return &SessionsReadHandler{svc: svc}
}

func (h *SessionsReadHandler) VerifySession(w http.ResponseWriter, r *http.Request) {
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
