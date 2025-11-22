package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"rgs/game"
	"rgs/sqlc"
)

type BetAggregate struct {
	queries *sqlc.Queries
	wallet  *WalletClient
}

func NewBetAggregate(q *sqlc.Queries, w *WalletClient) *BetAggregate {
	return &BetAggregate{
		queries: q,
		wallet:  w,
	}
}

type PlaceBetParams struct {
	OperatorID     int32
	PlayerID       int32
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

func (b *BetAggregate) settleWin(ctx context.Context, p PlaceBetParams, winAmount float64) (string, error) {
	creditKey := p.IdempotencyKey + "-win"

	ok, err := b.wallet.Credit(ctx, p.PlayerID, winAmount, creditKey)
	if err != nil || !ok {
		return "pending_settlement", nil
	}

	return "won", nil
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

	ok, err := b.wallet.Debit(ctx, p.PlayerID, p.Amount, p.IdempotencyKey)
	if err != nil {
		return sqlc.Round{}, sqlc.Bet{}, fmt.Errorf("wallet debit failed: %w", err)
	}
	if !ok {
		return sqlc.Round{}, sqlc.Bet{}, errors.New("insufficient funds")
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
	}

	if winAmount > 0 {
		status, _ = b.settleWin(ctx, p, winAmount)
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

	switch status {

	case "won":
		b.emitWebhookEvent(ctx, p.OperatorID, "settlement_success", map[string]interface{}{
			"bet_id":    bet.ID,
			"round_id":  round.ID,
			"amount":    winAmount,
			"status":    "won",
			"player_id": p.PlayerID,
		})

	case "pending_settlement":
		b.emitWebhookEvent(ctx, p.OperatorID, "settlement_pending", map[string]interface{}{
			"bet_id":    bet.ID,
			"round_id":  round.ID,
			"amount":    winAmount,
			"status":    "pending",
			"player_id": p.PlayerID,
		})

	case "lost":
		b.emitWebhookEvent(ctx, p.OperatorID, "settlement_lost", map[string]interface{}{
			"bet_id":    bet.ID,
			"round_id":  round.ID,
			"amount":    0,
			"status":    "lost",
			"player_id": p.PlayerID,
		})
	}

	if status == "pending_settlement" {
		_, _ = b.queries.InsertOutbox(ctx, sqlc.InsertOutboxParams{
			BetID:      bet.ID,
			OperatorID: p.OperatorID,
			PlayerID:   p.PlayerID,
			Amount:     winAmount,
		})
	}

	return round, bet, nil
}

func (b *BetAggregate) emitWebhookEvent(ctx context.Context, operatorID int32, eventType string, payload interface{}) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}

	_, _ = b.queries.InsertWebhookEvent(ctx, sqlc.InsertWebhookEventParams{
		OperatorID: sql.NullInt32{Int32: operatorID, Valid: true},
		EventType:  eventType,
		Payload:    raw,
	})
}
