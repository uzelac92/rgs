package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"rgs/game"
	"rgs/sqlc"
)

type BetAggregate struct {
	queries *sqlc.Queries
}

func NewBetAggregate(q *sqlc.Queries) *BetAggregate {
	return &BetAggregate{queries: q}
}

type PlaceBetParams struct {
	OperatorID     int32
	PlayerID       int32
	RoundID        int32
	Amount         float64
	IdempotencyKey string
}

func generateRandomSeed() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (b *BetAggregate) PlaceBet(ctx context.Context, p PlaceBetParams) (sqlc.Round, sqlc.Bet, error) {
	existing, err := b.queries.GetBetByIdempotency(ctx, sqlc.GetBetByIdempotencyParams{
		OperatorID:     p.OperatorID,
		IdempotencyKey: p.IdempotencyKey,
	})

	if err == nil {
		round, _ := b.queries.GetRound(ctx, existing.RoundID)
		return round, existing, nil
	}

	serverSeed, err := generateRandomSeed()
	if err != nil {
		return sqlc.Round{}, sqlc.Bet{}, err
	}

	clientSeed, err := generateRandomSeed()
	if err != nil {
		return sqlc.Round{}, sqlc.Bet{}, err
	}

	pf := game.GenerateOutcome(serverSeed, clientSeed)

	round, err := b.queries.CreateRound(ctx, sqlc.CreateRoundParams{
		OperatorID: p.OperatorID,
		PlayerID:   p.PlayerID,
		ServerSeed: serverSeed,
		ClientSeed: clientSeed,
		Outcome:    pf.Outcome,
	})
	if err != nil {
		return sqlc.Round{}, sqlc.Bet{}, err
	}

	winAmount := 0.0
	status := "lost"
	if pf.Outcome == 6 {
		winAmount = p.Amount * 5
		status = "won"
	}

	bet, err := b.queries.CreateBet(ctx, sqlc.CreateBetParams{
		OperatorID:     p.OperatorID,
		PlayerID:       p.PlayerID,
		RoundID:        round.ID,
		Amount:         p.Amount,
		Outcome:        pf.Outcome,
		WinAmount:      winAmount,
		Status:         status,
		IdempotencyKey: p.IdempotencyKey,
	})
	if err != nil {
		return sqlc.Round{}, sqlc.Bet{}, err
	}

	return round, bet, nil
}
