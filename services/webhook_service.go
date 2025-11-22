package services

import (
	"context"
	"rgs/sqlc"
)

type WebhookService struct {
	q *sqlc.Queries
}

func NewWebhookService(q *sqlc.Queries) *WebhookService {
	return &WebhookService{q: q}
}

func (s *WebhookService) RetryWebhook(ctx context.Context, id int32) error {
	_, err := s.q.GetWebhookEventByID(ctx, id)
	if err != nil {
		return err
	}

	return s.q.ResetWebhookForRetry(ctx, id)
}

func (s *WebhookService) ListWebhooks(
	ctx context.Context,
	operatorID int32,
	status *string,
) ([]sqlc.WebhookEvent, error) {
	if status == nil {
		return s.q.ListWebhooksByOperator(ctx, operatorID)
	}

	return s.q.ListWebhooksByOperatorStatus(ctx, sqlc.ListWebhooksByOperatorStatusParams{
		OperatorID: operatorID,
		Status:     *status,
	})
}
