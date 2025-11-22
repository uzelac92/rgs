package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"rgs/game"
	"rgs/observability"
	"rgs/sqlc"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type BetAggregate struct {
	queries    *sqlc.Queries
	db         *sql.DB
	wallet     *WalletClient
	bus        *EventBus
	compliance *ComplianceService
}

func NewBetAggregate(
	q *sqlc.Queries,
	wallet *WalletClient,
	bus *EventBus,
	db *sql.DB,
	compliance *ComplianceService,
) *BetAggregate {
	return &BetAggregate{
		queries:    q,
		db:         db,
		wallet:     wallet,
		bus:        bus,
		compliance: compliance,
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

func (b *BetAggregate) withTx(ctx context.Context, fn func(*sqlc.Queries) error) error {
	tx, err := b.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	qtx := b.queries.WithTx(tx)

	if err := fn(qtx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (b *BetAggregate) PlaceBet(ctx context.Context, p PlaceBetParams) (sqlc.Round, sqlc.Bet, error) {
	var round sqlc.Round
	var bet sqlc.Bet

	timer := prometheus.NewTimer(observability.BetSettlementDuration)
	defer timer.ObserveDuration()

	err := b.withTx(ctx, func(q *sqlc.Queries) error {
		existing, err := q.GetBetByIdempotency(ctx, sqlc.GetBetByIdempotencyParams{
			OperatorID:     p.OperatorID,
			IdempotencyKey: p.IdempotencyKey,
		})
		if err == nil {
			round, _ = q.GetRound(ctx, existing.RoundID)
			bet = existing
			return nil
		}

		player, err := b.queries.GetPlayerByID(ctx, p.PlayerID)
		if err != nil {
			return errors.New("player not found")
		}

		jurisdiction := player.Jurisdiction

		if b.compliance != nil {
			if err := b.compliance.Check(ctx, p.OperatorID, p.PlayerID, jurisdiction, p.Amount); err != nil {
				return err
			}
		}

		serverSeed, err := generateRandomSeed()
		if err != nil {
			return err
		}
		clientSeed, err := generateRandomSeed()
		if err != nil {
			return err
		}

		pf := game.GenerateOutcome(serverSeed, clientSeed)

		round, err = q.CreateRound(ctx, sqlc.CreateRoundParams{
			OperatorID: p.OperatorID,
			PlayerID:   p.PlayerID,
			ServerSeed: serverSeed,
			ClientSeed: clientSeed,
			Outcome:    pf.Outcome,
		})
		if err != nil {
			return err
		}

		if b.bus != nil {
			b.bus.Publish(SSEEvent{
				ID:         uuid.NewString(),
				OperatorID: p.OperatorID,
				EventType:  "round.finished",
				Data: map[string]any{
					"round_id":    round.ID,
					"player_id":   p.PlayerID,
					"server_seed": serverSeed,
					"client_seed": clientSeed,
					"outcome":     pf.Outcome,
				},
				CreatedAt: time.Now(),
			})
		}

		bet, err = q.CreateBet(ctx, sqlc.CreateBetParams{
			OperatorID:     p.OperatorID,
			PlayerID:       p.PlayerID,
			RoundID:        round.ID,
			Amount:         p.Amount,
			Outcome:        pf.Outcome,
			WinAmount:      0,
			Status:         "processing",
			IdempotencyKey: p.IdempotencyKey,
		})
		if err != nil {
			return err
		}

		observability.WalletDebitCalls.Inc()
		ok, errDebit := b.wallet.Debit(ctx, p.PlayerID, p.Amount, p.IdempotencyKey)
		if errDebit != nil || !ok {
			observability.WalletDebitFailures.Inc()
			errorMsg := "wallet debit failed"
			if errDebit != nil {
				errorMsg = errDebit.Error()
			}
			observability.Logger.Error("wallet debit failed", zap.Any("error", errorMsg))
			return errors.New("wallet debit failed")
		}

		if b.bus != nil {
			b.bus.Publish(SSEEvent{
				ID:         uuid.NewString(),
				OperatorID: p.OperatorID,
				EventType:  "wallet.debit",
				Data: map[string]any{
					"player_id": p.PlayerID,
					"amount":    p.Amount,
				},
				CreatedAt: time.Now(),
			})
		}

		winAmount := 0.0
		status := "lost"

		if pf.Outcome == 6 {
			winAmount = p.Amount * 5
			status = "won"
		}

		if status == "won" {
			creditKey := p.IdempotencyKey + "-win"

			okCredit, errCredit := b.wallet.Credit(ctx, p.PlayerID, winAmount, creditKey)
			if errCredit != nil || !okCredit {
				status = "pending_settlement"

				_, _ = q.InsertOutbox(ctx, sqlc.InsertOutboxParams{
					BetID:      bet.ID,
					OperatorID: p.OperatorID,
					PlayerID:   p.PlayerID,
					Amount:     winAmount,
				})
			}
		}

		bet, err = q.UpdateBetStatus(ctx, sqlc.UpdateBetStatusParams{
			ID:        bet.ID,
			Status:    status,
			WinAmount: winAmount,
		})
		if err != nil {
			return err
		}

		if b.bus != nil {
			eventType := "settlement.lost"
			if status == "won" {
				eventType = "settlement.won"
			} else if status == "pending_settlement" {
				eventType = "settlement.pending"
			}

			b.bus.Publish(SSEEvent{
				ID:         uuid.NewString(),
				OperatorID: p.OperatorID,
				EventType:  eventType,
				Data: map[string]any{
					"bet_id":    bet.ID,
					"round_id":  round.ID,
					"player_id": p.PlayerID,
					"amount":    winAmount,
					"status":    status,
				},
				CreatedAt: time.Now(),
			})
		}

		b.emitWebhookEvent(ctx, p.OperatorID, "bet_settled", map[string]any{
			"bet_id":    bet.ID,
			"round_id":  round.ID,
			"player_id": p.PlayerID,
			"amount":    winAmount,
			"status":    status,
		})

		return nil
	})

	if err != nil {
		return sqlc.Round{}, sqlc.Bet{}, err
	}

	return round, bet, nil
}

func (b *BetAggregate) emitWebhookEvent(ctx context.Context, operatorID int32, eventType string, payload interface{}) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return
	}

	_, _ = b.queries.InsertWebhookEvent(ctx, sqlc.InsertWebhookEventParams{
		OperatorID: operatorID,
		EventType:  eventType,
		Payload:    raw,
	})
}
