package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"rgs/sqlc"

	"github.com/go-chi/chi/v5"
)

type RoundsReadHandler struct {
	q *sqlc.Queries
}

func NewRoundsReadHandler(q *sqlc.Queries) *RoundsReadHandler {
	return &RoundsReadHandler{q: q}
}

func (h *RoundsReadHandler) GetRound(w http.ResponseWriter, r *http.Request) {
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

	round, err := h.q.GetRound(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "round not found", http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(round)
	if err != nil {
		log.Println("error encoding round", err)
	}
}
