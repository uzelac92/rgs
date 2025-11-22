package services

import (
	"context"
	"fmt"
	"rgs/observability"
	"rgs/sqlc"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OutboxWorker struct {
	queries *sqlc.Queries
	wallet  *WalletClient
	bus     *EventBus
}

func NewOutboxWorker(q *sqlc.Queries, wallet *WalletClient, bus *EventBus) *OutboxWorker {
	return &OutboxWorker{queries: q, wallet: wallet, bus: bus}
}

func (w *OutboxWorker) Start() {
	go func() {
		for {
			w.processPending()
			time.Sleep(3 * time.Second)
		}
	}()
}

func (w *OutboxWorker) processPending() {
	ctx := context.Background()

	events, err := w.queries.GetPendingOutbox(ctx)
	if err != nil {
		observability.Logger.Error("failed to fetch outbox", zap.Error(err))
		return
	}

	for _, e := range events {
		creditKey := fmt.Sprintf("bet-%d-retry-%d", e.BetID, e.ID)
		if w.bus != nil {
			w.bus.Publish(SSEEvent{
				ID:         uuid.NewString(),
				OperatorID: e.OperatorID,
				EventType:  "settlement.retry",
				Data: map[string]any{
					"bet_id":    e.BetID,
					"player_id": e.PlayerID,
					"amount":    e.Amount,
					"outbox_id": e.ID,
					"retry_key": creditKey,
				},
				CreatedAt: time.Now(),
			})
		}

		ok, errCredit := w.wallet.Credit(ctx, e.PlayerID, e.Amount, creditKey)
		if errCredit != nil || !ok {
			observability.Logger.Error("retry credit failed", zap.Error(err))

			errorMsg := "wallet declined"
			if errCredit != nil {
				errorMsg = errCredit.Error()
			}

			if w.bus != nil {
				w.bus.Publish(SSEEvent{
					ID:         uuid.NewString(),
					OperatorID: e.OperatorID,
					EventType:  "settlement.failed",
					Data: map[string]any{
						"bet_id":    e.BetID,
						"player_id": e.PlayerID,
						"amount":    e.Amount,
						"error":     errorMsg,
						"outbox_id": e.ID,
					},
					CreatedAt: time.Now(),
				})
			}

			continue
		}

		err = w.queries.MarkBetAsWon(ctx, e.BetID)
		if err != nil {
			observability.Logger.Error("failed to mark bet as won", zap.Error(err))
			continue
		}
		if w.bus != nil {
			w.bus.Publish(SSEEvent{
				ID:         uuid.NewString(),
				OperatorID: e.OperatorID,
				EventType:  "settlement.success",
				Data: map[string]any{
					"bet_id":    e.BetID,
					"player_id": e.PlayerID,
					"amount":    e.Amount,
					"outbox_id": e.ID,
				},
				CreatedAt: time.Now(),
			})
		}

		_, _ = w.queries.InsertWebhookEvent(ctx, sqlc.InsertWebhookEventParams{
			OperatorID: e.OperatorID,
			EventType:  "settlement_success",
			Payload: []byte(fmt.Sprintf(`{
						"bet_id": %d,
						"round_id": 0,
						"amount": %.2f,
						"status": "won",
						"player_id": %d
					}`, e.BetID, e.Amount, e.PlayerID)),
		})

		_, err = w.queries.MarkOutboxProcessed(ctx, e.ID)
		if err != nil {
			observability.Logger.Error("failed to mark outbox processed", zap.Error(err))
		}
	}
}
