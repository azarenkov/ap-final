package postgres

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"

	"notification-service/internal/domain"
)

type Repo struct{ db *sql.DB }

func New(db *sql.DB) *Repo { return &Repo{db: db} }

func (r *Repo) Insert(ctx context.Context, n *domain.Notification) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notifications (id, user_id, kind, subject, body, read, created_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		n.ID, n.UserID, n.Kind, n.Subject, n.Body, n.Read, n.CreatedAt)
	return err
}

func (r *Repo) GetByID(ctx context.Context, id string) (*domain.Notification, error) {
	var n domain.Notification
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, kind, subject, body, read, created_at
		 FROM notifications WHERE id = $1`, id,
	).Scan(&n.ID, &n.UserID, &n.Kind, &n.Subject, &n.Body, &n.Read, &n.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &n, err
}

func (r *Repo) ListByUser(ctx context.Context, userID string) ([]*domain.Notification, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, kind, subject, body, read, created_at
		 FROM notifications WHERE user_id = $1 ORDER BY created_at DESC LIMIT 200`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Kind, &n.Subject, &n.Body, &n.Read, &n.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &n)
	}
	return out, rows.Err()
}

func (r *Repo) MarkRead(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `UPDATE notifications SET read = TRUE WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}
