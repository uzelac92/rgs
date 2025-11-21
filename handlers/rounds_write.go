package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"rgs/middleware"
	"rgs/services"
)

type RoundsWriteHandler struct {
	svc *services.RoundsService
}

func NewRoundsWriteHandler(svc *services.RoundsService) *RoundsWriteHandler {
	return &RoundsWriteHandler{svc: svc}
}

type createRoundRequest struct {
	PlayerID int32 `json:"player_id"`
}

type createRoundResponse struct {
	RoundID    int32  `json:"round_id"`
	Outcome    int32  `json:"outcome"`
	ServerSeed string `json:"server_seed"`
	ClientSeed string `json:"client_seed"`
	Hash       string `json:"hash"`
}

func (h *RoundsWriteHandler) CreateRound(w http.ResponseWriter, r *http.Request) {
	var req createRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	op, ok := middleware.OperatorFromContext(r.Context())
	if !ok {
		http.Error(w, "operator missing", http.StatusUnauthorized)
		return
	}

	round, pf, err := h.svc.CreateRound(r.Context(), services.CreateRoundParams{
		OperatorID: op.ID,
		PlayerID:   req.PlayerID,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := createRoundResponse{
		RoundID:    round.ID,
		Outcome:    pf.Outcome,
		ServerSeed: pf.ServerSeed,
		ClientSeed: pf.ClientSeed,
		Hash:       pf.Hash,
	}

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Println("error encoding response:", err)
	}
}
