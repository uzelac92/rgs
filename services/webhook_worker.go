package services

import (
	"context"
	"database/sql"
	"log"
	"rgs/sqlc"
	"time"
)

type WebhookWorker struct {
	q *sqlc.Queries
}

func NewWebhookWorker(q *sqlc.Queries) *WebhookWorker {
	return &WebhookWorker{
		q: q,
	}
}

func (w *WebhookWorker) Start() {
	go func() {
		for {
			w.processPending()
			time.Sleep(3 * time.Second)
		}
	}()
}

func (w *WebhookWorker) processPending() {
	ctx := context.Background()

	events, err := w.q.GetPendingWebhookEvents(ctx)
	if err != nil {
		log.Println("failed to fetch pending webhook events:", err)
		return
	}

	for _, ev := range events {
		event, err := w.q.MarkWebhookProcessing(ctx, ev.ID)
		if err != nil {
			log.Println("failed to lock webhook event:", err)
			continue
		}

		if !event.OperatorID.Valid {
			_ = w.q.MarkWebhookFailed(ctx, sqlc.MarkWebhookFailedParams{
				ID: event.ID,
				ErrorMessage: sql.NullString{
					String: "operator_id null",
					Valid:  true,
				},
			})
			continue
		}

		operator, err := w.q.GetOperatorByID(ctx, event.OperatorID.Int32)
		if err != nil || operator.WebhookUrl == "" {
			_ = w.q.MarkWebhookFailed(ctx, sqlc.MarkWebhookFailedParams{
				ID: event.ID,
				ErrorMessage: sql.NullString{
					String: "operator not found or no webhook_url",
					Valid:  true,
				},
			})
			continue
		}

		if time.Since(event.CreatedAt) > time.Minute {
			_ = w.q.MarkWebhookFailed(ctx, sqlc.MarkWebhookFailedParams{
				ID: event.ID,
				ErrorMessage: sql.NullString{
					String: "retry window exceeded",
					Valid:  true,
				},
			})
			continue
		}

		client := NewWebhookClient(operator.WebhookSecret)
		err = client.Send(ctx, operator.WebhookUrl, event.Payload)
		if err == nil {
			_ = w.q.MarkWebhookCompleted(ctx, event.ID)
			continue
		}

		delay := nextRetryDelay(event.Retries)
		next := time.Now().Add(delay)

		_, _ = w.q.UpdateWebhookRetry(ctx, sqlc.UpdateWebhookRetryParams{
			ID:          event.ID,
			NextRetryAt: next,
			ErrorMessage: sql.NullString{
				String: err.Error(),
				Valid:  true,
			},
		})
	}
}

func nextRetryDelay(retries int32) time.Duration {
	switch retries {
	case 0:
		return 0
	case 1:
		return 5 * time.Second
	case 2:
		return 10 * time.Second
	case 3:
		return 20 * time.Second
	default:
		return 25 * time.Second
	}
}
