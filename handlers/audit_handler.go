package handlers

import (
	"encoding/json"
	"net/http"
	"rgs/middleware"
	"rgs/services"
	"strconv"
)

type AuditHandler struct {
	svc *services.ComplianceService
}

func NewAuditHandler(svc *services.ComplianceService) *AuditHandler {
	return &AuditHandler{svc: svc}
}

func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	operator, ok := middleware.OperatorFromContext(r.Context())
	if !ok {
		http.Error(w, "operator missing", http.StatusUnauthorized)
		return
	}

	playerIDParam := r.URL.Query().Get("player_id")
	var playerID *int32
	if playerIDParam != "" {
		val, err := strconv.Atoi(playerIDParam)
		if err == nil {
			tmp := int32(val)
			playerID = &tmp
		}
	}

	limitParam := r.URL.Query().Get("limit")
	offsetParam := r.URL.Query().Get("offset")

	limit := int32(50)
	offset := int32(0)

	if limitParam != "" {
		if v, err := strconv.Atoi(limitParam); err == nil {
			limit = int32(v)
		}
	}
	if offsetParam != "" {
		if v, err := strconv.Atoi(offsetParam); err == nil {
			offset = int32(v)
		}
	}

	logs, err := h.svc.ListLogs(r.Context(), operator.ID, playerID, limit, offset)
	if err != nil {
		http.Error(w, "failed to fetch audit logs", http.StatusInternalServerError)
		return
	}

	_ = json.NewEncoder(w).Encode(logs)
}
