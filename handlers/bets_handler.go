package handlers

import (
	"encoding/json"
	"net/http"
	"rgs/middleware"
	"rgs/observability"
	"rgs/services"

	"go.uber.org/zap"
)

type BetsHandler struct {
	agg *services.BetAggregate
}

func NewBetsHandler(agg *services.BetAggregate) *BetsHandler {
	return &BetsHandler{agg: agg}
}

type placeBetRequest struct {
	RoundID        int32   `json:"round_id"`
	PlayerID       int32   `json:"player_id"`
	Amount         float64 `json:"amount"`
	IdempotencyKey string  `json:"idempotency_key"`
}

func (h *BetsHandler) PlaceBet(w http.ResponseWriter, r *http.Request) {
	var req placeBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		observability.Logger.Error("error decoding betting request", zap.Error(err))
		return
	}

	operator, ok := middleware.OperatorFromContext(r.Context())
	if !ok {
		http.Error(w, "missing operator", http.StatusUnauthorized)
		return
	}

	observability.BetsPlaced.Inc()
	round, bet, err := h.agg.PlaceBet(r.Context(), services.PlaceBetParams{
		OperatorID:     operator.ID,
		PlayerID:       req.PlayerID,
		Amount:         req.Amount,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	observability.Logger.Info("bet placed",
		zap.Int32("operator_id", operator.ID),
		zap.Float64("amount", req.Amount),
	)

	resp := struct {
		BetID   int32 `json:"bet_id"`
		RoundID int32 `json:"round_id"`
	}{
		BetID:   bet.ID,
		RoundID: round.ID,
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		observability.Logger.Error("error encoding placeBet", zap.Error(err))
		return
	}
}
