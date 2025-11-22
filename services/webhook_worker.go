package services

import (
	"context"
	"database/sql"
	"rgs/observability"
	"rgs/sqlc"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WebhookWorker struct {
	queries *sqlc.Queries
	bus     *EventBus
}

func NewWebhookWorker(q *sqlc.Queries, bus *EventBus) *WebhookWorker {
	return &WebhookWorker{
		queries: q,
		bus:     bus,
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

	events, err := w.queries.GetPendingWebhookEvents(ctx)
	if err != nil {
		observability.Logger.Error("failed to fetch pending webhook events:", zap.Error(err))
		return
	}

	for _, ev := range events {
		event, err := w.queries.MarkWebhookProcessing(ctx, ev.ID)
		if err != nil {
			observability.Logger.Error("failed to lock webhook event:", zap.Error(err))
			continue
		}

		operator, err := w.queries.GetOperatorByID(ctx, event.OperatorID)
		if err != nil || operator.WebhookUrl == "" {
			_ = w.queries.MarkWebhookFailed(ctx, sqlc.MarkWebhookFailedParams{
				ID: event.ID,
				ErrorMessage: sql.NullString{
					String: "operator not found or no webhook_url",
					Valid:  true,
				},
			})
			continue
		}

		if time.Since(event.CreatedAt) > time.Minute {
			_ = w.queries.MarkWebhookFailed(ctx, sqlc.MarkWebhookFailedParams{
				ID: event.ID,
				ErrorMessage: sql.NullString{
					String: "retry window exceeded",
					Valid:  true,
				},
			})
			continue
		}

		if w.bus != nil {
			w.bus.Publish(SSEEvent{
				ID:         uuid.NewString(),
				OperatorID: event.OperatorID,
				EventType:  "webhook.retry",
				Data: map[string]any{
					"event_id":   event.ID,
					"event_type": event.EventType,
					"retries":    event.Retries,
				},
				CreatedAt: time.Now(),
			})
		}

		client := NewWebhookClient(operator.WebhookSecret)
		err = client.Send(ctx, operator.WebhookUrl, event.Payload)
		if err == nil {
			_ = w.queries.MarkWebhookCompleted(ctx, event.ID)
			continue
		}

		delay := nextRetryDelay(event.Retries)
		next := time.Now().Add(delay)

		_, _ = w.queries.UpdateWebhookRetry(ctx, sqlc.UpdateWebhookRetryParams{
			ID:          event.ID,
			NextRetryAt: next,
			ErrorMessage: sql.NullString{
				String: err.Error(),
				Valid:  true,
			},
		})

		if w.bus != nil {
			w.bus.Publish(SSEEvent{
				ID:         uuid.NewString(),
				OperatorID: event.OperatorID,
				EventType:  "webhook.failed",
				Data: map[string]any{
					"event_id":   event.ID,
					"event_type": event.EventType,
					"error":      err.Error(),
					"retry_in":   delay.Seconds(),
				},
				CreatedAt: time.Now(),
			})
		}
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
