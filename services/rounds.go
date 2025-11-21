package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"rgs/game"
	"rgs/sqlc"
)

type RoundsService struct {
	queries *sqlc.Queries
}

func NewRoundsService(q *sqlc.Queries) *RoundsService {
	return &RoundsService{queries: q}
}

func generateRandomSeed() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

type CreateRoundParams struct {
	OperatorID int32
	PlayerID   int32
}

func (s *RoundsService) CreateRound(ctx context.Context, p CreateRoundParams) (sqlc.Round, game.ProvablyFairResult, error) {
	// Step 1 — generate both seeds
	serverSeed, err := generateRandomSeed()
	if err != nil {
		return sqlc.Round{}, game.ProvablyFairResult{}, err
	}

	clientSeed, err := generateRandomSeed()
	if err != nil {
		return sqlc.Round{}, game.ProvablyFairResult{}, err
	}

	// Step 2 — generate provably fair outcome
	pf := game.GenerateOutcome(serverSeed, clientSeed)

	// Step 3 — save in database
	round, err := s.queries.CreateRound(ctx, sqlc.CreateRoundParams{
		OperatorID: p.OperatorID,
		PlayerID:   p.PlayerID,
		ServerSeed: serverSeed,
		ClientSeed: clientSeed,
		Outcome:    pf.Outcome,
	})
	if err != nil {
		return sqlc.Round{}, game.ProvablyFairResult{}, err
	}

	return round, pf, nil
}
