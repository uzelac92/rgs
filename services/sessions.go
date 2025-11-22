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
	queries    *sqlc.Queries
	bus        *EventBus
	compliance *ComplianceService
}

func NewSessionsService(
	q *sqlc.Queries,
	bus *EventBus,
	comp *ComplianceService,
) *SessionsService {
	return &SessionsService{queries: q, bus: bus, compliance: comp}
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
	if err := s.compliance.CheckJurisdiction(ctx, p.OperatorID, p.Jurisdiction); err != nil {
		return sqlc.Session{}, err
	}

	id := uuid.New()

	player, err := s.getOrCreatePlayer(ctx, p.OperatorID, p.ExternalPlayerID, p.Jurisdiction)
	if err != nil {
		return sqlc.Session{}, err
	}

	launchToken, err := generateSecureToken()
	if err != nil {
		return sqlc.Session{}, err
	}

	session, err := s.queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:          id,
		OperatorID:  p.OperatorID,
		PlayerID:    player.ID,
		LaunchToken: launchToken,
		ExpiresAt:   time.Now().Add(p.TTL),
	})
	if err != nil {
		return sqlc.Session{}, err
	}

	s.compliance.Log(ctx, p.OperatorID, &player.ID, "session.launch", map[string]any{
		"external_player_id": p.ExternalPlayerID,
		"jurisdiction":       p.Jurisdiction,
		"session_id":         session.ID.String(),
	})

	s.bus.Publish(SSEEvent{
		ID:         uuid.NewString(),
		OperatorID: p.OperatorID,
		EventType:  "session.launched",
		Data: map[string]any{
			"session_id": session.ID,
			"player_id":  player.ID,
			"expires_at": session.ExpiresAt,
		},
		CreatedAt: time.Now(),
	})

	return session, nil
}

func (s *SessionsService) VerifySession(ctx context.Context, token string, operatorID int32) (sqlc.Session, error) {
	session, err := s.queries.VerifySessionByToken(ctx, sqlc.VerifySessionByTokenParams{
		LaunchToken: token,
		OperatorID:  operatorID,
	})
	if err != nil {
		return sqlc.Session{}, err
	}

	s.bus.Publish(SSEEvent{
		ID:         uuid.NewString(),
		OperatorID: operatorID,
		EventType:  "session.verified",
		Data:       session,
		CreatedAt:  time.Now(),
	})

	return session, nil
}

func (s *SessionsService) RevokeSession(ctx context.Context, id uuid.UUID, operatorID int32) error {
	err := s.queries.RevokeSession(ctx, id)
	if err != nil {
		return err
	}

	s.bus.Publish(SSEEvent{
		ID:         uuid.NewString(),
		OperatorID: operatorID,
		EventType:  "session.revoked",
		Data: map[string]any{
			"session_id": id.String(),
		},
		CreatedAt: time.Now(),
	})

	return nil
}
