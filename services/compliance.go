package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"rgs/sqlc"
)

type ComplianceService struct {
	queries *sqlc.Queries
}

func NewComplianceService(q *sqlc.Queries) *ComplianceService {
	return &ComplianceService{queries: q}
}

func (s *ComplianceService) Log(
	ctx context.Context,
	operatorID int32,
	playerID *int32,
	action string,
	details any,
) {
	raw, err := json.Marshal(details)
	if err != nil {
		return
	}

	pid := int32(0)
	if playerID != nil {
		pid = *playerID
	}

	_ = s.queries.InsertAuditLog(ctx, sqlc.InsertAuditLogParams{
		OperatorID: operatorID,
		PlayerID: sql.NullInt32{
			Int32: pid,
			Valid: playerID != nil,
		},
		Action:  action,
		Details: raw,
	})
}

func (s *ComplianceService) CheckJurisdiction(
	ctx context.Context,
	operatorID int32,
	jurisdiction string,
) error {
	limits, err := s.queries.GetOperatorLimits(ctx, operatorID)
	if err != nil {
		return nil
	}

	if len(limits.AllowedJurisdictions) == 0 {
		return nil
	}

	for _, j := range limits.AllowedJurisdictions {
		if j == jurisdiction {
			return nil
		}
	}

	return errors.New("jurisdiction not allowed")
}

func (s *ComplianceService) CheckMaxBet(ctx context.Context, operatorID int32, amount float64) error {
	limits, err := s.queries.GetOperatorLimits(ctx, operatorID)
	if err != nil {
		return nil
	}

	if limits.MaxBet == 0 {
		return nil
	}

	if amount > limits.MaxBet {
		return errors.New("bet exceeds operator max bet limit")
	}

	return nil
}

func (s *ComplianceService) Check(
	ctx context.Context,
	operatorID int32,
	playerID int32,
	jurisdiction string,
	amount float64,
) error {
	if err := s.CheckJurisdiction(ctx, operatorID, jurisdiction); err != nil {
		s.Log(ctx, operatorID, &playerID, "compliance.jurisdiction_block", map[string]any{
			"jurisdiction": jurisdiction,
		})
		return err
	}

	if err := s.CheckMaxBet(ctx, operatorID, amount); err != nil {
		s.Log(ctx, operatorID, &playerID, "compliance.max_bet_block", map[string]any{
			"amount": amount,
		})
		return err
	}

	return nil
}

func (s *ComplianceService) ListLogs(
	ctx context.Context,
	operatorID int32,
	playerID *int32,
	limit, offset int32,
) ([]sqlc.AuditLog, error) {
	if playerID == nil {
		return s.queries.ListAuditLogsByOperator(ctx, sqlc.ListAuditLogsByOperatorParams{
			OperatorID: operatorID,
			Limit:      limit,
			Offset:     offset,
		})
	}

	return s.queries.ListAuditLogsByOperatorPlayer(ctx, sqlc.ListAuditLogsByOperatorPlayerParams{
		OperatorID: operatorID,
		PlayerID: sql.NullInt32{
			Int32: *playerID,
			Valid: true,
		},
		Limit:  limit,
		Offset: offset,
	})
}
