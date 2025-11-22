package main

import "sync"

type Store struct {
	balances map[int32]float64
	idem     map[string]bool
	mu       sync.Mutex
}

func NewStore() *Store {
	s := &Store{
		balances: make(map[int32]float64),
		idem:     make(map[string]bool),
	}

	// default balance for player with id 1
	s.balances[1] = 100000

	return s
}

func (s *Store) CheckIdempotent(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.idem[key] {
		return true
	}

	s.idem[key] = true
	return false
}

func (s *Store) GetBalance(player int32) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.balances[player]
}

func (s *Store) Debit(player int32, amount float64) (ok bool, newBalance float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	bal := s.balances[player]
	if bal < amount {
		return false, bal
	}

	newBal := bal - amount
	s.balances[player] = newBal
	return true, newBal
}

func (s *Store) Credit(player int32, amount float64) (bool, float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.balances[player] += amount
	return true, s.balances[player]
}
