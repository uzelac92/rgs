package services

import (
	"context"
	"rgs/sqlc"
)

type WebhookService struct {
	queries *sqlc.Queries
}

func NewWebhookService(q *sqlc.Queries) *WebhookService {
	return &WebhookService{queries: q}
}

func (s *WebhookService) RetryWebhook(ctx context.Context, id int32) error {
	_, err := s.queries.GetWebhookEventByID(ctx, id)
	if err != nil {
		return err
	}

	return s.queries.ResetWebhookForRetry(ctx, id)
}

func (s *WebhookService) ListWebhooks(
	ctx context.Context,
	operatorID int32,
	status *string,
) ([]sqlc.WebhookEvent, error) {
	if status == nil {
		return s.queries.ListWebhooksByOperator(ctx, operatorID)
	}

	return s.queries.ListWebhooksByOperatorStatus(ctx, sqlc.ListWebhooksByOperatorStatusParams{
		OperatorID: operatorID,
		Status:     *status,
	})
}
