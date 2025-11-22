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
	http.HandleFunc("/wallet/debit", s.handlers.Debit)
	http.HandleFunc("/wallet/credit", s.handlers.Credit)

	log.Println("Wallet mock running on :9000")
	log.Fatal(http.ListenAndServe(":9000", nil))
}
