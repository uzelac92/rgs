package handlers

import (
	"encoding/json"
	"net/http"
	"rgs/middleware"
	"rgs/observability"
	"rgs/services"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type WebhookHandler struct {
	svc *services.WebhookService
}

func NewWebhookHandler(svc *services.WebhookService) *WebhookHandler {
	return &WebhookHandler{svc: svc}
}

func (h *WebhookHandler) RetryWebhook(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		observability.Logger.Error("retry webhook invalid", zap.Any("id", idStr), zap.Error(err))
		return
	}

	err = h.svc.RetryWebhook(r.Context(), int32(id64))
	if err != nil {
		http.Error(w, "event not found", http.StatusNotFound)
		observability.Logger.Error("retry webhook failed", zap.Error(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("retry scheduled"))
}

func (h *WebhookHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
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

	events, err := h.svc.ListWebhooks(r.Context(), operator.ID, status)
	if err != nil {
		http.Error(w, "failed to load webhooks", http.StatusInternalServerError)
		observability.Logger.Error("failed to load webhooks", zap.Error(err))
		return
	}

	err = json.NewEncoder(w).Encode(events)
	if err != nil {
		observability.Logger.Error("failed to encode webhooks", zap.Error(err))
		return
	}
}
