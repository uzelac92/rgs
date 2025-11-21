package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"rgs/middleware"
	"rgs/services"
)

type BetsWriteHandler struct {
	agg *services.BetAggregate
}

func NewBetsWriteHandler(agg *services.BetAggregate) *BetsWriteHandler {
	return &BetsWriteHandler{agg: agg}
}

type placeBetRequest struct {
	RoundID        int32   `json:"round_id"`
	PlayerID       int32   `json:"player_id"`
	Amount         float64 `json:"amount"`
	IdempotencyKey string  `json:"idempotency_key"`
}

func (h *BetsWriteHandler) PlaceBet(w http.ResponseWriter, r *http.Request) {
	var req placeBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	operator, ok := middleware.OperatorFromContext(r.Context())
	if !ok {
		http.Error(w, "missing operator", http.StatusUnauthorized)
		return
	}

	bet, err := h.agg.PlaceBet(r.Context(), services.PlaceBetParams{
		OperatorID:     operator.ID,
		PlayerID:       req.PlayerID,
		RoundID:        req.RoundID,
		Amount:         req.Amount,
		IdempotencyKey: req.IdempotencyKey,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = json.NewEncoder(w).Encode(bet)
	if err != nil {
		log.Println("error encoding placeBet", err)
	}
}
