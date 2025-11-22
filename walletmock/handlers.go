package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Handlers struct {
	store *Store
}

func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

func (h *Handlers) Debit(w http.ResponseWriter, r *http.Request) {
	var req WalletRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("error decoding request", err)
		return
	}

	if !ValidSignature(req.PlayerID, req.Amount, req.RequestID, req.Signature) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	if h.store.CheckIdempotent(req.RequestID) {
		err := json.NewEncoder(w).Encode(WalletResponse{
			Success: true,
			Balance: h.store.GetBalance(req.PlayerID),
		})
		if err != nil {
			log.Println("error encoding response", err)
		}
		return
	}

	ok, newBalance := h.store.Debit(req.PlayerID, req.Amount)

	err = json.NewEncoder(w).Encode(WalletResponse{
		Success: ok,
		Balance: newBalance,
	})
	if err != nil {
		log.Println("error encoding wallet response", err)
		return
	}
}

func (h *Handlers) Credit(w http.ResponseWriter, r *http.Request) {
	var req WalletRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("error decoding request", err)
		return
	}

	if !ValidSignature(req.PlayerID, req.Amount, req.RequestID, req.Signature) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	if h.store.CheckIdempotent(req.RequestID) {
		err := json.NewEncoder(w).Encode(WalletResponse{
			Success: true,
			Balance: h.store.GetBalance(req.PlayerID),
		})
		if err != nil {
			log.Println("error encoding wallet response", err)
		}
		return
	}

	ok, newBalance := h.store.Credit(req.PlayerID, req.Amount)

	err = json.NewEncoder(w).Encode(WalletResponse{
		Success: ok,
		Balance: newBalance,
	})
	if err != nil {
		log.Println("error encoding wallet response", err)
		return
	}
}
