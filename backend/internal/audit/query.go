package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LogEntry struct {
	ID          uuid.UUID       `json:"id"`
	ActorUserID *uuid.UUID      `json:"actor_user_id,omitempty"`
	ActorEmail  *string         `json:"actor_email,omitempty"`
	Action      string          `json:"action"`
	EntityType  string          `json:"entity_type"`
	EntityID    *uuid.UUID      `json:"entity_id,omitempty"`
	RequestID   *uuid.UUID      `json:"request_id,omitempty"`
	OldValues   json.RawMessage `json:"old_values,omitempty"`
	NewValues   json.RawMessage `json:"new_values,omitempty"`
	IPAddress   *string         `json:"ip_address,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}

type ListFilter struct {
	Action       string
	EntityType   string
	ActorUserID  *uuid.UUID
	DateFrom     *time.Time
	DateTo       *time.Time
	Limit        int
	Offset       int
}

func (s *Service) List(ctx context.Context, f ListFilter) ([]LogEntry, int, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	if f.Limit > 5000 {
		f.Limit = 5000
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	where := []string{"1=1"}
	args := []any{}
	n := 1
	if a := strings.TrimSpace(f.Action); a != "" {
		where = append(where, fmt.Sprintf("al.action = $%d", n))
		args = append(args, a)
		n++
	}
	if et := strings.TrimSpace(f.EntityType); et != "" {
		where = append(where, fmt.Sprintf("al.entity_type = $%d", n))
		args = append(args, et)
		n++
	}
	if f.ActorUserID != nil {
		where = append(where, fmt.Sprintf("al.actor_user_id = $%d", n))
		args = append(args, *f.ActorUserID)
		n++
	}
	if f.DateFrom != nil {
		where = append(where, fmt.Sprintf("al.created_at >= $%d", n))
		args = append(args, *f.DateFrom)
		n++
	}
	if f.DateTo != nil {
		where = append(where, fmt.Sprintf("al.created_at < $%d", n))
		args = append(args, *f.DateTo)
		n++
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM audit_logs al WHERE `+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := fmt.Sprintf(`
		SELECT al.id, al.actor_user_id, u.email, al.action, al.entity_type, al.entity_id, al.request_id,
		       al.old_values, al.new_values, al.ip_address, al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON u.id = al.actor_user_id
		WHERE %s
		ORDER BY al.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, n, n+1)
	args = append(args, f.Limit, f.Offset)
	rows, err := s.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var items []LogEntry
	for rows.Next() {
		var e LogEntry
		if err := rows.Scan(&e.ID, &e.ActorUserID, &e.ActorEmail, &e.Action, &e.EntityType, &e.EntityID, &e.RequestID, &e.OldValues, &e.NewValues, &e.IPAddress, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, rows.Err()
}
