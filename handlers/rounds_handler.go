package handlers

import (
	"encoding/json"
	"net/http"
	"rgs/observability"
	"strconv"

	"rgs/sqlc"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type RoundsHandler struct {
	queries *sqlc.Queries
}

func NewRoundsHandler(q *sqlc.Queries) *RoundsHandler {
	return &RoundsHandler{queries: q}
}

func (h *RoundsHandler) GetRound(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "missing round id", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid round id", http.StatusBadRequest)
		return
	}

	round, err := h.queries.GetRound(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "round not found", http.StatusNotFound)
		observability.Logger.Error("round not found", zap.Error(err))
		return
	}

	err = json.NewEncoder(w).Encode(round)
	if err != nil {
		observability.Logger.Error("error encoding round", zap.Error(err))
		return
	}
}
