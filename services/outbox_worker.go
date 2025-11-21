package services

import (
	"context"
	"fmt"
	"log"
	"rgs/sqlc"
	"time"
)

type OutboxWorker struct {
	q      *sqlc.Queries
	wallet *WalletClient
}

func NewOutboxWorker(q *sqlc.Queries, wallet *WalletClient) *OutboxWorker {
	return &OutboxWorker{q: q, wallet: wallet}
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

	events, err := w.q.GetPendingOutbox(ctx)
	if err != nil {
		log.Println("failed to fetch outbox:", err)
		return
	}

	for _, e := range events {
		creditKey := fmt.Sprintf("bet-%d-retry-%d", e.BetID, e.ID)

		ok, err := w.wallet.Credit(ctx, e.PlayerID, e.Amount, creditKey)
		if err != nil || !ok {
			log.Println("retry credit failed:", err)
			continue
		}

		err = w.q.MarkBetAsWon(ctx, e.BetID)
		if err != nil {
			log.Println("failed to mark bet as won:", err)
			continue
		}

		_, err = w.q.MarkOutboxProcessed(ctx, e.ID)
		if err != nil {
			log.Println("failed to mark outbox processed:", err)
		}
	}
}
