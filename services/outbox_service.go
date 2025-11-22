package services

import (
	"context"
	"database/sql"
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

	var processed sql.NullBool

	switch *status {
	case "processed", "completed":
		processed = sql.NullBool{Bool: true, Valid: true}

	case "pending", "retrying":
		processed = sql.NullBool{Bool: false, Valid: true}

	default:
		return s.queries.ListOutboxByOperator(ctx, operatorID)
	}

	return s.queries.ListOutboxByOperatorStatus(
		ctx,
		sqlc.ListOutboxByOperatorStatusParams{
			OperatorID: operatorID,
			Processed:  processed,
		},
	)
}
