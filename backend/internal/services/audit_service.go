package services

import (
	"context"
	"database/sql"

	"corebank/backend/internal/models"
)

type AuditLogService struct {
	db *sql.DB
}

func NewAuditLogService(db *sql.DB) *AuditLogService {
	return &AuditLogService{db: db}
}

// execer is satisfied by both *sql.DB and *sql.Tx. Accepting the interface
// lets callers write an audit row either standalone or as part of the same
// database transaction as the banking operation it describes — an audit
// log for a transfer that never happened would be worse than no log at all.
type execer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (s *AuditLogService) Record(ctx context.Context, ex execer, action, actorType string, actorID *int64, entity string, entityID *int64, description string) error {
	_, err := ex.ExecContext(ctx, `
		INSERT INTO audit_logs (action, actor_type, actor_id, entity, entity_id, description)
		VALUES (?, ?, ?, ?, ?, ?)`,
		action, actorType, actorID, entity, entityID, description,
	)
	return err
}

func (s *AuditLogService) List(ctx context.Context, limit int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, action, actor_type, actor_id, entity, entity_id, description, created_at
		FROM audit_logs ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.AuditLog
	for rows.Next() {
		var a models.AuditLog
		var actorID, entityID sql.NullInt64
		if err := rows.Scan(&a.ID, &a.Action, &a.ActorType, &actorID, &a.Entity, &entityID, &a.Description, &a.CreatedAt); err != nil {
			return nil, err
		}
		if actorID.Valid {
			v := actorID.Int64
			a.ActorID = &v
		}
		if entityID.Valid {
			v := entityID.Int64
			a.EntityID = &v
		}
		out = append(out, a)
	}
	return out, rows.Err()
}
