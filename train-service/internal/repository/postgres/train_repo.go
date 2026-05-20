package postgres

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"train-service/internal/domain"
)

type TrainRepo struct {
	db *sql.DB
}

func NewTrainRepo(db *sql.DB) *TrainRepo {
	return &TrainRepo{db: db}
}

func (r *TrainRepo) Insert(ctx context.Context, t *domain.Train) error {
	const q = `
		INSERT INTO trains (id, code, name, route_id, departure_time, arrival_time,
		                    total_seats, available_seats, price_cents, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`
	_, err := r.db.ExecContext(ctx, q,
		t.ID, t.Code, t.Name, t.RouteID, t.DepartureTime, t.ArrivalTime,
		t.TotalSeats, t.AvailableSeats, t.PriceCents, t.Status, t.CreatedAt, t.UpdatedAt,
	)
	return err
}

func (r *TrainRepo) GetByID(ctx context.Context, id string) (*domain.Train, error) {
	const q = `
		SELECT id, code, name, route_id, departure_time, arrival_time,
		       total_seats, available_seats, price_cents, status, created_at, updated_at
		FROM trains WHERE id = $1
	`
	var t domain.Train
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&t.ID, &t.Code, &t.Name, &t.RouteID, &t.DepartureTime, &t.ArrivalTime,
		&t.TotalSeats, &t.AvailableSeats, &t.PriceCents, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrTrainNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TrainRepo) Update(ctx context.Context, t *domain.Train) error {
	const q = `
		UPDATE trains
		SET name = $2, departure_time = $3, arrival_time = $4,
		    price_cents = $5, status = $6, updated_at = $7
		WHERE id = $1
	`
	res, err := r.db.ExecContext(ctx, q,
		t.ID, t.Name, t.DepartureTime, t.ArrivalTime, t.PriceCents, t.Status, time.Now().UTC(),
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrTrainNotFound
	}
	return nil
}

func (r *TrainRepo) UpdateAvailableSeats(ctx context.Context, id string, delta int32) (int32, error) {
	const q = `
		UPDATE trains
		SET available_seats = available_seats + $2, updated_at = NOW()
		WHERE id = $1
		RETURNING available_seats
	`
	var available int32
	if err := r.db.QueryRowContext(ctx, q, id, delta).Scan(&available); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrTrainNotFound
		}
		if strings.Contains(err.Error(), "available_seats") {
			return 0, domain.ErrNotEnoughSeats
		}
		return 0, err
	}
	return available, nil
}

func (r *TrainRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM trains WHERE id = $1", id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrTrainNotFound
	}
	return nil
}

func (r *TrainRepo) Search(ctx context.Context, f *domain.SearchFilter) ([]*domain.Train, int32, error) {
	f.Normalize()

	args := []any{}
	conds := []string{"TRUE"}
	idx := 1
	add := func(cond string, val any) {
		conds = append(conds, strings.Replace(cond, "?", "$"+itoa(idx), 1))
		args = append(args, val)
		idx++
	}

	if f.Origin != "" {
		add("r.origin = ?", f.Origin)
	}
	if f.Destination != "" {
		add("r.destination = ?", f.Destination)
	}
	if f.DepartureAfter != nil {
		add("t.departure_time >= ?", *f.DepartureAfter)
	}
	if f.DepartureBefore != nil {
		add("t.departure_time <= ?", *f.DepartureBefore)
	}

	where := strings.Join(conds, " AND ")

	countQ := "SELECT COUNT(*) FROM trains t JOIN routes r ON r.id = t.route_id WHERE " + where
	var total int32
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ := "SELECT t.id, t.code, t.name, t.route_id, t.departure_time, t.arrival_time," +
		" t.total_seats, t.available_seats, t.price_cents, t.status, t.created_at, t.updated_at " +
		"FROM trains t JOIN routes r ON r.id = t.route_id WHERE " + where +
		" ORDER BY t.departure_time ASC LIMIT $" + itoa(idx) + " OFFSET $" + itoa(idx+1)
	args = append(args, f.PageSize, f.Offset())

	rows, err := r.db.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]*domain.Train, 0, f.PageSize)
	for rows.Next() {
		var t domain.Train
		if err := rows.Scan(
			&t.ID, &t.Code, &t.Name, &t.RouteID, &t.DepartureTime, &t.ArrivalTime,
			&t.TotalSeats, &t.AvailableSeats, &t.PriceCents, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		out = append(out, &t)
	}
	return out, total, rows.Err()
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := false
	if i < 0 {
		neg = true
		i = -i
	}
	var b [16]byte
	pos := len(b)
	for i > 0 {
		pos--
		b[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		b[pos] = '-'
	}
	return string(b[pos:])
}
