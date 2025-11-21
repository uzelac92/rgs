package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"rgs/sqlc"

	"github.com/google/uuid"
)

type SessionsService struct {
	queries *sqlc.Queries
}

func NewSessionsService(q *sqlc.Queries) *SessionsService {
	return &SessionsService{queries: q}
}

type LaunchSessionParams struct {
	OperatorID       int32
	ExternalPlayerID string
	Jurisdiction     string
	TTL              time.Duration
}

func generateSecureToken() (string, error) {
	b := make([]byte, 32) // 256 bits of entropy
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (s *SessionsService) getOrCreatePlayer(
	ctx context.Context,
	operatorID int32,
	externalID string,
	jurisdiction string,
) (sqlc.Player, error) {
	player, err := s.queries.GetPlayer(ctx, sqlc.GetPlayerParams{
		OperatorID:       operatorID,
		ExternalPlayerID: externalID,
	})

	if err == nil {
		return player, nil
	}

	return s.queries.CreatePlayer(ctx, sqlc.CreatePlayerParams{
		OperatorID:       operatorID,
		ExternalPlayerID: externalID,
		Jurisdiction:     jurisdiction,
	})
}

func (s *SessionsService) LaunchSession(ctx context.Context, p LaunchSessionParams) (sqlc.Session, error) {
	id := uuid.New()

	player, err := s.getOrCreatePlayer(ctx, p.OperatorID, p.ExternalPlayerID, p.Jurisdiction)
	if err != nil {
		return sqlc.Session{}, err
	}

	launchToken, err := generateSecureToken()
	if err != nil {
		return sqlc.Session{}, err
	}

	return s.queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:          id,
		OperatorID:  p.OperatorID,
		PlayerID:    player.ID,
		LaunchToken: launchToken,
		ExpiresAt:   time.Now().Add(p.TTL),
	})
}

func (s *SessionsService) VerifySession(ctx context.Context, token string, operatorID int32) (sqlc.Session, error) {
	return s.queries.VerifySessionByToken(ctx, sqlc.VerifySessionByTokenParams{
		LaunchToken: token,
		OperatorID:  operatorID,
	})
}

func (s *SessionsService) RevokeSession(ctx context.Context, id uuid.UUID) error {
	return s.queries.RevokeSession(ctx, id)
}
