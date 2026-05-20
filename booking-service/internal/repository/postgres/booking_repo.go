package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/lib/pq"

	"booking-service/internal/domain"
)

type BookingRepo struct{ db *sql.DB }

func NewBookingRepo(db *sql.DB) *BookingRepo { return &BookingRepo{db: db} }

func (r *BookingRepo) Insert(ctx context.Context, b *domain.Booking) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO bookings (id, user_id, train_id, seat_count, amount_cents, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		b.ID, b.UserID, b.TrainID, b.SeatCount, b.AmountCents, b.Status, b.CreatedAt, b.UpdatedAt)
	return err
}

func (r *BookingRepo) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	var b domain.Booking
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, train_id, seat_count, amount_cents, status, created_at, updated_at
		 FROM bookings WHERE id = $1`, id,
	).Scan(&b.ID, &b.UserID, &b.TrainID, &b.SeatCount, &b.AmountCents, &b.Status, &b.CreatedAt, &b.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrBookingNotFound
	}
	return &b, err
}

func (r *BookingRepo) UpdateStatus(ctx context.Context, id, status string) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE bookings SET status = $2, updated_at = NOW() WHERE id = $1`, id, status)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrBookingNotFound
	}
	return nil
}

func (r *BookingRepo) UpdateAmount(ctx context.Context, id string, amount int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE bookings SET amount_cents = $2, updated_at = NOW() WHERE id = $1`, id, amount)
	return err
}

func (r *BookingRepo) ListByUser(ctx context.Context, userID string) ([]*domain.Booking, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, train_id, seat_count, amount_cents, status, created_at, updated_at
		 FROM bookings WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	return scanList(rows)
}

func (r *BookingRepo) ListPage(ctx context.Context, page, size int32) ([]*domain.Booking, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, train_id, seat_count, amount_cents, status, created_at, updated_at
		 FROM bookings ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		size, (page-1)*size)
	if err != nil {
		return nil, err
	}
	return scanList(rows)
}

func (r *BookingRepo) InsertTicket(ctx context.Context, ticketID, bookingID, code string) (time.Time, error) {
	issued := time.Now().UTC()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO tickets (id, booking_id, code, issued_at) VALUES ($1,$2,$3,$4)
		 ON CONFLICT (id) DO NOTHING`,
		ticketID, bookingID, code, issued)
	return issued, err
}

func (r *BookingRepo) RefundInTx(ctx context.Context, id, actor string) (prevStatus string, amount int64, err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return "", 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = tx.QueryRowContext(ctx,
		`SELECT status, amount_cents FROM bookings WHERE id = $1 FOR UPDATE`, id,
	).Scan(&prevStatus, &amount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", 0, domain.ErrBookingNotFound
		}
		return "", 0, err
	}

	if prevStatus != domain.StatusConfirmed && prevStatus != domain.StatusCancelled {
		return prevStatus, 0, domain.ErrIllegalTransition
	}

	if _, err = tx.ExecContext(ctx,
		`UPDATE bookings SET status = $2, updated_at = NOW() WHERE id = $1`, id, domain.StatusRefunded,
	); err != nil {
		return prevStatus, 0, err
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO booking_audit (booking_id, action, old_status, new_status, actor)
		 VALUES ($1, $2, $3, $4, $5)`,
		id, "REFUND", prevStatus, domain.StatusRefunded, actor,
	); err != nil {
		return prevStatus, 0, err
	}

	if err = tx.Commit(); err != nil {
		return prevStatus, 0, err
	}
	return prevStatus, amount, nil
}

func scanList(rows *sql.Rows) ([]*domain.Booking, error) {
	defer rows.Close()
	var out []*domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.TrainID, &b.SeatCount, &b.AmountCents, &b.Status, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &b)
	}
	return out, rows.Err()
}
