package main

import (
	"log"
	"net/http"
)

type Server struct {
	store    *Store
	handlers *Handlers
}

func NewServer() *Server {
	store := NewStore()
	handlers := NewHandlers(store)

	return &Server{
		store:    store,
		handlers: handlers,
	}
}

func (s *Server) Run() {
	mux := http.NewServeMux()

	mux.HandleFunc("/wallet/debit", s.handlers.Debit)
	mux.HandleFunc("/wallet/credit", s.handlers.Credit)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	log.Println("Wallet mock running on :9000")
	log.Fatal(http.ListenAndServe(":9000", mux))
}
