package services

import (
	"context"
	"rgs/sqlc"
)

type OutboxService struct {
	queries *sqlc.Queries
}

func NewOutboxService(q *sqlc.Queries) *OutboxService {
	return &OutboxService{queries: q}
}

func (s *OutboxService) ListOutbox(
	ctx context.Context,
	operatorID int32,
	status *string,
) ([]sqlc.Outbox, error) {
	if status == nil {
		return s.queries.ListOutboxByOperator(ctx, operatorID)
	}

	var processedBool bool

	switch *status {
	case "processed", "completed":
		processedBool = true

	case "pending", "retrying":
		processedBool = false

	default:
		return s.queries.ListOutboxByOperator(ctx, operatorID)
	}

	return s.queries.ListOutboxByOperatorStatus(
		ctx,
		sqlc.ListOutboxByOperatorStatusParams{
			OperatorID: operatorID,
			Processed:  processedBool,
		},
	)
}
