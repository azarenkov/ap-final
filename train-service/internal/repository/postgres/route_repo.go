package postgres

import (
	"context"
	"database/sql"
	"errors"

	"train-service/internal/domain"
)

type RouteRepo struct {
	db *sql.DB
}

func NewRouteRepo(db *sql.DB) *RouteRepo {
	return &RouteRepo{db: db}
}

func (r *RouteRepo) Insert(ctx context.Context, x *domain.Route) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO routes (id, origin, destination, distance_km, estimated_minutes)
		 VALUES ($1,$2,$3,$4,$5)`,
		x.ID, x.Origin, x.Destination, x.DistanceKm, x.EstimatedMinutes,
	)
	return err
}

func (r *RouteRepo) GetByID(ctx context.Context, id string) (*domain.Route, error) {
	var x domain.Route
	err := r.db.QueryRowContext(ctx,
		`SELECT id, origin, destination, distance_km, estimated_minutes FROM routes WHERE id = $1`, id,
	).Scan(&x.ID, &x.Origin, &x.Destination, &x.DistanceKm, &x.EstimatedMinutes)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrRouteNotFound
	}
	if err != nil {
		return nil, err
	}
	return &x, nil
}

func (r *RouteRepo) Update(ctx context.Context, x *domain.Route) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE routes SET distance_km = $2, estimated_minutes = $3 WHERE id = $1`,
		x.ID, x.DistanceKm, x.EstimatedMinutes,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrRouteNotFound
	}
	return nil
}

func (r *RouteRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM routes WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrRouteNotFound
	}
	return nil
}
