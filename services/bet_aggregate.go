package services

import (
	"context"
	"errors"
	"rgs/sqlc"
)

var (
	ErrDuplicateBet        = errors.New("duplicate bet: idempotency key already used")
	ErrInvalidRound        = errors.New("round does not exist")
	ErrWrongOperator       = errors.New("round does not belong to operator")
	ErrRoundPlayerMismatch = errors.New("round does not belong to this player")
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

func (b *BetAggregate) PlaceBet(ctx context.Context, p PlaceBetParams) (sqlc.Bet, error) {
	existing, err := b.queries.GetBetByIdempotency(ctx, sqlc.GetBetByIdempotencyParams{
		OperatorID:     p.OperatorID,
		IdempotencyKey: p.IdempotencyKey,
	})

	if err == nil {
		return existing, ErrDuplicateBet
	}

	round, err := b.queries.GetRound(ctx, p.RoundID)
	if err != nil {
		return sqlc.Bet{}, ErrInvalidRound
	}

	if round.OperatorID != p.OperatorID {
		return sqlc.Bet{}, ErrWrongOperator
	}

	if round.PlayerID != p.PlayerID {
		return sqlc.Bet{}, ErrRoundPlayerMismatch
	}

	winAmount := 0.0
	status := "lost"
	if round.Outcome == 6 {
		winAmount = p.Amount * 2
		status = "won"
	}

	bet, err := b.queries.CreateBet(ctx, sqlc.CreateBetParams{
		OperatorID:     p.OperatorID,
		PlayerID:       p.PlayerID,
		RoundID:        p.RoundID,
		Amount:         p.Amount,
		Outcome:        round.Outcome,
		WinAmount:      winAmount,
		Status:         status,
		IdempotencyKey: p.IdempotencyKey,
	})

	if err != nil {
		return sqlc.Bet{}, err
	}

	return bet, nil
}
